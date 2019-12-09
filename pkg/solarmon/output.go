package solarmon

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// const MaxInfluxQSize = 1440
// const MaxPostsPerFlush = 200
const MaxInfluxQSize = 120
const MaxPostsPerFlush = 12

type Output interface {
	WriteRegisters([]*Register) error
	Write(string) error
}

func GetOutputs(cfg *Config) map[string]Output {
	var defaultOut Output
	outputs := make(map[string]Output)
	if cfg.Once {
		outputs["stdout"] = NewStdOutput()
		defaultOut = outputs["stdout"]
	} else {
		header := fmt.Sprintf(
			"start=%s %s",
			cfg.StartTime.Format("2006-01-02 15:04:05"),
			cfg.Version,
		)
		outputs["http"] = NewHTTPOutput(cfg.HTTPListen, header)
		defaultOut = outputs["http"]
	}

	if cfg.Influxdb != "" {
		outputs["influx"] = NewInfluxOutput(
			cfg.Influxdb, cfg.InfluxTags, cfg.InfluxDry, defaultOut,
		)
	}

	return outputs
}

func WriteToAllOutputs(outputs map[string]Output, data interface{}) {
	for _, output := range outputs {
		switch v := data.(type) {
		case []*Register:
			output.WriteRegisters(v)
		case string:
			output.Write(v)
		}
	}
}

type StdOutput struct{}

func NewStdOutput() *StdOutput {
	o := &StdOutput{}
	return o
}

func (o *StdOutput) WriteRegisters(data []*Register) error {
	for _, register := range data {
		fmt.Println(register.String())
	}
	return nil
}

func (o *StdOutput) Write(data string) error {
	fmt.Printf("%s ## %s\n", time.Now().Format("2006-01-02 15:04:05"), data)
	return nil
}

type HTTPOutput struct {
	listen    string
	maxBufLen int
	header    string
	mutex     *sync.Mutex
	shortTxt  string
	longTxt   string
}

func NewHTTPOutput(listen string, header string) *HTTPOutput {
	o := &HTTPOutput{}
	o.maxBufLen = 1 * 1024 * 1024
	o.listen = listen
	o.header = header
	o.mutex = &sync.Mutex{}
	o.start()
	return o
}

func (o *HTTPOutput) WriteRegisters(data []*Register) error {
	var shortTxt, longTxt string
	for _, register := range data {
		r := register.String()
		longTxt += r + "\n"

		if strings.Contains(r, "**") {
			shortTxt += r + "\n"
		}
	}

	o.updateContents(shortTxt, longTxt)
	return nil
}

func (o *HTTPOutput) Write(data string) error {
	var shortTxt string
	longTxt := fmt.Sprintf("%s ## %s", time.Now().Format("2006-01-02 15:04:05"), data)
	if strings.Contains(data, "**") {
		shortTxt = longTxt
	}

	o.updateContents(shortTxt, longTxt)
	return nil
}

func (o *HTTPOutput) updateContents(bodyShort, bodyLong string) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if len(o.longTxt) > o.maxBufLen {
		o.longTxt = ""
	}

	if len(o.shortTxt) > o.maxBufLen {
		o.shortTxt = ""
	}

	o.longTxt = bodyLong + "\n" + o.longTxt
	if len(bodyShort) > 0 {
		o.shortTxt = bodyShort + "\n" + o.shortTxt
	}
	return nil
}

func (o *HTTPOutput) start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		o.mutex.Lock()
		defer o.mutex.Unlock()
		header := fmt.Sprintf("now=%s %s\n\n", time.Now().Format("2006-01-02 15:04:05"), o.header)

		keys, ok := r.URL.Query()["long"]

		if !ok || len(keys[0]) < 1 {
			fmt.Fprintf(w, header+o.shortTxt)
		} else {
			fmt.Fprintf(w, header+o.longTxt)
		}
	})

	go http.ListenAndServe(o.listen, nil)
}

type InfluxOutput struct {
	out        Output
	uri        string
	globalTags string
	q          [][]string
	dryRun     bool
	httpClient *http.Client
}

func NewInfluxOutput(uri string, globalTags string, dryRun bool, defaultOut Output) *InfluxOutput {
	o := &InfluxOutput{
		uri:        uri,
		globalTags: globalTags,
		dryRun:     dryRun,
		out:        defaultOut,
		q:          make([][]string, 0, 30),
	}

	transport := &http.Transport{
		MaxIdleConns:        5,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 15 * time.Second,
	}

	o.httpClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	return o
}

func (o *InfluxOutput) WriteRegisters(data []*Register) error {
	queries := o.prepareQueries(data)

	if len(queries) == 0 {
		o.out.Write("ERR Influxdb: no queries to exec\n")
		return nil
	}

	if len(o.q) >= MaxInfluxQSize {
		o.out.Write(fmt.Sprintf("ERR Influxdb: max qsize[%d]\n", MaxInfluxQSize))
		o.q = o.q[1:]
	}

	o.q = append(o.q, queries)

	startQSize := len(o.q)
	okReqs := 0
	for tries := 1; len(o.q) > 0 && tries <= MaxPostsPerFlush; tries++ {
		err := o.executeQueries(o.q[0])
		if err != nil {
			errTxt := strings.Replace(err.Error(), o.uri, "http://endpoint", 1)
			o.out.Write(fmt.Sprintf("ERR Influxdb query: qsize=%d req=%d %s\n", len(o.q), tries, errTxt))
			break
		} else {
			o.q = o.q[1:]
			okReqs += 1
		}
	}

	if startQSize > 1 && okReqs > 0 {
		o.out.Write(fmt.Sprintf("OK Influxdb query: qsize=%d finished_reqs=%d\n", len(o.q), okReqs))
	}
	return nil
}

func (o *InfluxOutput) Write(data string) error {
	return nil
}

func (o *InfluxOutput) prepareQueries(registers []*Register) []string {
	measurements := make(map[string]map[string][]string)
	for _, register := range registers {
		//cpu_load_short,host=server01,region=us-west value=0.64,val2=111 1434055562000000000

		if _, ok := measurements[register.MName]; !ok {
			measurements[register.MName] = make(map[string][]string)
		}

		loc, _ := time.LoadLocation("UTC")
		ts := time.Now().In(loc)

		register.Mutex.Lock()

		if register.TsType == "none" || register.lastErr != nil {
			register.Mutex.Unlock()
			continue
		}

		switch register.TsType {
		case "5m":
			min := ts.Minute() - (ts.Minute() % 5)
			ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), min, 0, 0, time.UTC)
		case "-1h":
			ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour()-1, 0, 0, 0, time.UTC)
		case "1h":
			ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, time.UTC)
		case "-1d":
			ts = time.Date(ts.Year(), ts.Month(), ts.Day()-1, 5, 0, 0, 0, time.UTC)
		case "1d":
			ts = time.Date(ts.Year(), ts.Month(), ts.Day(), 5, 0, 0, 0, time.UTC)
		case "1m":
			ts = time.Date(ts.Year(), ts.Month(), 1, 5, 0, 0, 0, time.UTC)
		case "-1m":
			ts = time.Date(ts.Year(), time.Month(int(ts.Month())-1), 1, 5, 0, 0, 0, time.UTC)
		case "1y":
			ts = time.Date(ts.Year(), time.Month(1), 1, 5, 0, 0, 0, time.UTC)
		case "-1y":
			ts = time.Date(ts.Year()-1, time.Month(1), 1, 5, 0, 0, 0, time.UTC)
		case "inf":
			ts = time.Unix(0, 0)
		}

		fieldName := strings.Trim(register.name, "* ")
		fieldName = strings.Split(fieldName, "^")[0]

		registerValue := fmt.Sprintf("%s=%s", fieldName, register.value)
		tsvalue := fmt.Sprintf("%v", ts.UnixNano())

		measurements[register.MName][tsvalue] = append(
			measurements[register.MName][tsvalue], registerValue,
		)

		register.Mutex.Unlock()
	}

	queries := []string{}

	for measurement, timestamps := range measurements {
		header := fmt.Sprintf("%s,%s ", measurement, o.globalTags)
		q := ""

		for ts, values := range timestamps {
			q += fmt.Sprintf("%s %s %s\n", header, strings.Join(values, ","), ts)
		}

		if len(q) > 0 {
			queries = append(queries, q)
		}
	}

	return queries
}

func (o *InfluxOutput) executeQueries(queries []string) error {
	body := []byte(strings.Join(queries, ""))

	if o.dryRun {
		o.out.Write("\ncurl '" + o.uri + "' --data-binary '" + string(body) + "'\n")
		return nil
	}

	req, err := http.NewRequest("POST", o.uri, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}
