package solarmon

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func NewConfig(InfluxEndpoint, InfluxTags, BuildTime, GitCommit string) *Config {

	config := &Config{}

	flag.StringVar(&config.ModbusParity, "prty", "N", "Parity")
	flag.IntVar(&config.StopBits, "sb", 1, "Stop Bits")
	flag.DurationVar(&config.Timeout, "t", 5*time.Second, "Max secs to wait for single read to finish")
	flag.IntVar(&config.DataBits, "db", 8, "Data Bits")
	flag.IntVar(&config.BaudRate, "br", 19200, "Baud Rate")
	flag.UintVar(&config.SlaveID, "slaveId", 15, "Slave ID")
	flag.StringVar(&config.TTYFile, "tty", "/dev/ttyUSB0", "TTY device file/name")
	flag.StringVar(&config.ReadRegistersFromFile, "rfile", "", "File with registers list")
	flag.BoolVar(&config.Once, "once", false, "Run only once and exit")
	flag.DurationVar(&config.ReadInterval, "interval", 10*time.Second, "Seconds to wait between reads")

	flag.BoolVar(&config.InfluxDry, "influxDry", false, "Just print influx queries on stdout")

	if InfluxEndpoint != "" {
		flag.Usage = func() { fmt.Println("Solarmon ... ab@val-energy.com") }

		if InfluxEndpoint == "ENV" {
			d := os.Getenv("VAL_CUST_NAME")
			u := os.Getenv("VAL_INFLUX_USER")
			p := os.Getenv("VAL_INFLUX_PASS")
			InfluxEndpoint = fmt.Sprintf(
				"https://%s.%s:8086/write?db=db0&u=%s&p=%s", d, defaultInflxDomain, u, p,
			)
		}

		config.Influxdb = InfluxEndpoint
		config.ReadInterval = 300 * time.Second
	} else {
		flag.StringVar(&config.Influxdb, "influxdb", "", "Influx db: http://localhost:8086/write?db=dbmae&u=user&p=pass")
	}

	if InfluxTags != "" {
		if InfluxTags == "ENV" {
			InfluxTags = os.Getenv("VAL_INFLUX_TAGS")
		}
		config.InfluxTags = InfluxTags
	} else {
		flag.StringVar(&config.InfluxTags, "influxTags", "loc=1,type=1,inverter=ktl33", "Tags to write with every measurement")
	}

	flag.BoolVar(&config.NMode, "nightmode", true, "Sleep during the night")
	flag.IntVar(&config.NModeStart, "nightmodeStart", 22, "Night starts at")
	flag.IntVar(&config.NModeEnd, "nightmodeEnd", 5, "Night ends at")
	flag.DurationVar(&config.NModeSleepInterval, "nightmodeSleep", 5*time.Minute, "See every nightmodeSleep minutes if the night has ended")
	flag.StringVar(&config.HTTPListen, "HTTPListen", ":8090", "HTTP listen addr")

	showVersion := flag.Bool("v", false, "show version")

	flag.Parse()

	config.Version = fmt.Sprintf("build=%s git=%s", BuildTime, GitCommit)
	config.StartTime = time.Now()

	if config.NModeStart < 19 || config.NModeStart > 23 {
		fmt.Printf("%s: Nightmode can only start between 19h and 23h \n", config.StartTime.Format("2006-01-02 15:04:05"))
		os.Exit(1)
	}

	if *showVersion {
		fmt.Printf("%s: %s \n", config.StartTime.Format("2006-01-02 15:04:05"), config.Version)
		os.Exit(0)
	}

	config.ReadRegistersFromCli = flag.Args()

	if len(config.ReadRegistersFromCli) == 0 && config.ReadRegistersFromFile == "" {
		config.ReadRegistersFromCli = strings.Split(defaultRfile, "\n")
	}

	config.DefaultMName = "solar"
	config.DefaultTsType = "now"

	return config
}

type Config struct {
	ModbusParity          string
	StopBits              int
	Timeout               time.Duration
	DataBits              int
	BaudRate              int
	SlaveID               uint
	InvertorType          string
	TTYFile               string
	ReadRegistersFromCli  []string
	ReadRegistersFromFile string
	ReadInterval          time.Duration
	Once                  bool
	NMode                 bool
	NModeStart            int
	NModeEnd              int
	NModeSleepInterval    time.Duration
	Influxdb              string
	InfluxTags            string
	InfluxDry             bool
	HTTPListen            string
	Version               string
	StartTime             time.Time
	DefaultTsType         string
	DefaultMName          string
}

const defaultInflxDomain = "mon.val-energy.com"

// DefaultRfile when no rfile is provided - SUN2000-30KTL-M3 MODBUS Interface Definitions-2021-08-12
const defaultRfile = `32080:2:active_power**:1000:I32:kW
32082:2:reactive_power**:1000:I32:kVar
32064:2:input_power**:1000:I32:kW
32078:2:peak_power**:1000:I32:kW
32114:2:eday**:100:U32:kWh:1d:eday
32106:2:etotal**:100:U32:kWh:inf:etotal
32087:1:temp**:10:I16:C
32086:1:efficiency:100:U16:%%
32088:1:iResistance:1000:U16:Mohm
32016:1:Upv1:10:I16:V
32017:1:Ipv1:100:I16:A
32018:1:Upv2:10:I16:V
32019:1:Ipv2:100:I16:A
32020:1:Upv3:10:I16:V
32021:1:Ipv3:100:I16:A
32022:1:Upv4:10:I16:V
32023:1:Ipv4:100:I16:A
32024:1:Upv5:10:I16:V
32025:1:Ipv5:100:I16:A
32026:1:Upv6:10:I16:V
32027:1:Ipv6:100:I16:A
32028:1:Upv7:10:I16:V
32029:1:Ipv7:100:I16:A
32030:1:Upv8:10:I16:V
32031:1:Ipv8:100:I16:A
32000:1:state1:1:U16:_
32002:1:state2:1:U16:_
32003:2:state3:1:U32:_
32008:1:alarm1:1:U16:_
32009:1:alarm2:1:U16:_
32010:1:alarm3:1:U16:_
32089:1:status:1:U16:_
32090:1:faultcode:1:U16:_
32091:2:startupts:2:U32:_
32093:2:shutdownts:2:U32:_
40000:2:ctimets:2:U32:_
32066:1:Uab:10:U16:V
32067:1:Ubc:10:U16:V
32068:1:Uca:10:U16:V
32069:1:Ua:10:U16:V
32070:1:Ub:10:U16:V
32071:1:Uc:10:U16:V
32072:2:Ia:1000:I32:A
32074:2:Ib:1000:I32:A
32076:2:Ic:1000:I32:A
32085:1:freq:100:U16:Hz
32084:1:power_factor:1000:I16:_:none
`
