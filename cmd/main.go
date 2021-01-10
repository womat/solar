package main

import (
	"github.com/womat/debug"
	"os"
	"os/signal"
	"solar/pkg/solar"
	"syscall"
)

func main() {
	initConfig()
	initWebService()

	debug.SetDebug(Config.Debug.File, Config.Debug.Flag)

	Measurements = CurrentMeasurements{
		Values: solar.Values{
			State:          solar.Off,
			SolarPumpState: solar.Off,
			BrinePumpState: solar.Off,
		},
	}

	if err := Measurements.load(Config.DataFile); err != nil {
		debug.ErrorLog.Printf("can't open data file: %v\n", err)
		os.Exit(1)
		return
	}

	results := &solarPumpRuntime{
		handler: solar.New(),
		current: &Measurements,
		last: lastValues{
			state:          solar.Off,
			brinePumpState: solar.Off,
			solarPumpState: solar.Off,
		},
	}

	results.handler.SetUVS232URL(Config.UVS232URL)

	go results.serveCalc(Config.DataCollectionInterval)
	go results.serveSave(Config.DataFile, Config.BackupInterval)

	// capture exit signals to ensure resources are released on exit.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// wait for am os.Interrupt signal (CTRL C)
	sig := <-quit
	debug.InfoLog.Printf("Got %s signal. Aborting...\n", sig)
	_ = Measurements.save(Config.DataFile)
	os.Exit(1)
}
