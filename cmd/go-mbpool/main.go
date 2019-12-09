package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tolivb/go-mbpool/pkg/solarmon"
)

var GitCommit string
var BuildTime string
var InfluxEndpoint string
var InfluxTags string

func main() {
	config := solarmon.NewConfig(InfluxEndpoint, InfluxTags, BuildTime, GitCommit)

	fmt.Printf("%s: starting Solarmon ...\n", config.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("%s: VER=%s\n", config.StartTime.Format("2006-01-02 15:04:05"), config.Version)
	fmt.Printf("%s: HTTP=%s\n", config.StartTime.Format("2006-01-02 15:04:05"), config.HTTPListen)
	fmt.Printf("%s: TAGS=%s\n", config.StartTime.Format("2006-01-02 15:04:05"), config.InfluxTags)
	fmt.Printf("%s: INTERVAL=%s\n", config.StartTime.Format("2006-01-02 15:04:05"), config.ReadInterval)
	fmt.Printf(
		"%s: NIGHTMODE=%v, from=%v, to=%v\n",
		config.StartTime.Format("2006-01-02 15:04:05"),
		config.NMode,
		config.NModeStart,
		config.NModeEnd,
	)

	registers, err := solarmon.GetRegistersToRead(config)

	if len(registers) < 1 {
		fmt.Fprintf(os.Stderr, "%s: ERR: %s\n", time.Now().String(), err)
		os.Exit(1)
	}

	modbusRTU, err := solarmon.NewModbusRTU(config)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: ERR: %s\n", time.Now().String(), err)
		os.Exit(2)
	}

	outputs := solarmon.GetOutputs(config)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Printf("%s: stopping Solarmon ...\n\n", time.Now().Format("2006-01-02 15:04:05"))
		modbusRTU.Close()
		os.Exit(0)
	}()

	for {
		t0 := time.Now()
		// No need of empty values during the night
		if config.NMode && (t0.Hour() >= config.NModeStart || t0.Hour() <= config.NModeEnd) && !config.Once {
			solarmon.WriteToAllOutputs(
				outputs,
				fmt.Sprintf("Nightmode from %d to %d", config.NModeStart, config.NModeEnd),
			)
			time.Sleep(config.NModeSleepInterval)
			continue
		}

		for _, r := range registers {
			r.ReadHR(modbusRTU)
		}

		t1 := time.Now()
		solarmon.WriteToAllOutputs(outputs, registers)
		t2 := time.Now()

		duration := fmt.Sprintf("%s, %s; **\n", fmt.Sprint(t1.Sub(t0)), fmt.Sprint(t2.Sub(t1)))
		solarmon.WriteToAllOutputs(outputs, duration)

		if config.Once {
			sigs <- syscall.SIGINT
		}

		time.Sleep(config.ReadInterval)
	}
}
