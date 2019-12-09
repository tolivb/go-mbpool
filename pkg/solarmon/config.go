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

	flag.BoolVar(&config.NMode, "nightmode", false, "Sleep during the night")
	flag.IntVar(&config.NModeStart, "nightmodeStart", 22, "Night starts at")
	flag.IntVar(&config.NModeEnd, "nightmodeEnd", 5, "Night ends at")
	flag.DurationVar(&config.NModeSleepInterval, "nightmodeSleep", 5*time.Minute, "See every nightmodeSleep minutes if the night has ended")
	flag.StringVar(&config.HTTPListen, "HTTPListen", ":8090", "HTTP listen addr")

	showVersion := flag.Bool("v", false, "show version")

	flag.Parse()

	config.Version = fmt.Sprintf("build=%s git=%s", BuildTime, GitCommit)
	config.StartTime = time.Now()

	if *showVersion {
		fmt.Printf("%s: %s \n", config.StartTime.Format("2006-01-02 15:04:05"), config.Version)
		os.Exit(0)
	}

	config.ReadRegistersFromCli = flag.Args()

	if len(config.ReadRegistersFromCli) == 0 && config.ReadRegistersFromFile == "" {
		config.ReadRegistersFromCli = strings.Split(defaultRfile, "\n")
	}

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
}

const defaultInflxDomain = "mon.val-energy.com"

// DefaultRfile when no rfile is provided - SUN2000 36,42,33KTL-A MODBUS Interface Definitions-20170731
const defaultRfile = `32290:2:active_power**:1000:I32:kW
32292:2:reactive_power**:1000:I32:kVar
32294:2:input_power**:1000:U32:kW
32298:2:ehour**:100:U32:kWh:1h:ehour
32345:2:ehour^p:100:U32:kWh:-1h:ehour
32300:2:eday**:100:U32:kWh:1d:eday
32349:2:eday^p:100:U32:kWh:-1d:eday
32302:2:emonth**:100:U32:kWh:1m:emonth
32353:2:emonth^p:100:U32:kWh:-1m:emonth
32304:2:eyear:100:U32:kWh:1y:eyear
32357:2:eyear^p:100:U32:kWh:-1y:eyear
32306:2:etotal**:100:U32:kWh:inf:etotal
32286:1:temp**:10:I16:C
32262:1:Upv1:10:I16:V
32263:1:Ipv1:10:I16:A
32264:1:Upv2:10:I16:V
32265:1:Ipv2:10:I16:A
32266:1:Upv3:10:I16:V
32267:1:Ipv3:10:I16:A
32268:1:Upv4:10:I16:V
32269:1:Ipv4:10:I16:A
32270:1:Upv5:10:I16:V
32271:1:Ipv5:10:I16:A
32272:1:Upv6:10:I16:V
32273:1:Ipv6:10:I16:A
32314:1:Upv7:10:I16:V
32315:1:Ipv7:10:I16:A
32316:1:Upv8:10:I16:V
32317:1:Ipv8:10:I16:A
32285:1:efficiency:100:U16:%%
32322:1:ongrid**:1:U16:_
32323:1:iResistance:1000:U16:Mohm
33022:2:inP_MPPT1:1000:U32:kW
33024:2:inP_MPPT2:1000:U32:kW
33026:2:inP_MPPT3:1000:U32:kW
33070:2:inP_MPPT4:1000:U32:kW
32274:1:Uab:10:U16:V
32275:1:Ubc:10:U16:V
32276:1:Uca:10:U16:V
32277:1:Ua:10:U16:V
32278:1:Ub:10:U16:V
32279:1:Uc:10:U16:V
32280:1:Ia:10:U16:A
32281:1:Ib:10:U16:A
32282:1:Ic:10:U16:A
32283:1:freq:100:U16:Hz
32284:1:power_factor:1000:U16:_:none
32287:1:inv_status:1:U16:_:none
32288:2:peak_power**:1000:I32:kW
32319:1:s1:1:U16:_
32320:1:s2:1:U16:_
32321:1:s3:1:U16:_
50000:1:a1:1:U16:_
50001:1:a2:1:U16:_
50002:1:a3:1:U16:_
50003:1:a4:1:U16:_
50004:1:a5:1:U16:_
50005:1:a6:1:U16:_
50006:1:a7:1:U16:_
50007:1:a8:1:U16:_
50008:1:a9:1:U16:_
50009:1:a10:1:U16:_
50010:1:a11:1:U16:_
50011:1:a12:1:U16:_
50012:1:a13:1:U16:_
50013:1:a14:1:U16:_
50014:1:a15:1:U16:_
50015:1:a16:1:U16:_
50016:1:a17:1:U16:_
`

// const defaultRfile = `32080:2:active_power**:1000:I32:kW
// 32064:2:input_power**:1000:I32:kW
// 32082:2:reactive_power**:1000:I32:kVar
// 32078:2:peak_power**:1000:I32:kW
// 32114:2:eday**:100:U32:kWh:1d:eday
// 32106:2:etotal**:100:U32:kWh:inf:etotal
// 32087:1:temp**:10:I16:C
// 32016:1:Upv1:10:I16:V
// 32017:1:Ipv1:100:I16:A
// 32018:1:Upv2:10:I16:V
// 32019:1:Ipv2:100:I16:A
// 32020:1:Upv3:10:I16:V
// 32021:1:Ipv3:100:I16:A
// 32022:1:Upv4:10:I16:V
// 32023:1:Ipv4:100:I16:A
// 32086:1:efficiency:100:U16:%%
// 32003:2:ongrid**:1:U32:_
// 32088:1:iResistance:1000:U16:Mohm
// 32066:1:Uab:10:U16:V
// 32067:1:Ubc:10:U16:V
// 32068:1:Uca:10:U16:V
// 32069:1:Ua:10:U16:V
// 32070:1:Ub:10:U16:V
// 32071:1:Uc:10:U16:V
// 32072:2:Ia:1000:I32:A
// 32074:2:Ib:1000:I32:A
// 32076:2:Ic:1000:I32:A
// 32085:1:freq:100:U16:Hz
// 32084:1:power_factor:1000:I16:_:none
// 32089:1:inv_status:1:U16:_:none
// 32000:1:s1:1:U16:_
// 32002:1:s2:1:U16:_
// 32003:2:s3:1:U32:_
// 32008:1:a1:1:U16:_
// 32009:1:a2:1:U16:_
// 32010:1:a3:1:U16:_
// `
