package main

import (
	"io"
	"solar/pkg/solar"
	"sync"
	"time"
)

// VERSION holds the version information with the following logic in mind
//  1 ... fixed
//  0 ... year 2020, 1->year 2021, etc.
//  7 ... month of year (7=July)
//  the date format after the + is always the first of the month
//
// VERSION differs from semantic versioning as described in https://semver.org/
// but we keep the correct syntax.
//TODO: increase version number to 1.0.1+2020xxyy
const VERSION = "0.0.1+20201227"
const MODULE = "solar"

type DebugConf struct {
	File io.WriteCloser
	Flag int
}

type WebserverConf struct {
	Port        int
	Webservices map[string]bool
}

type Configuration struct {
	DataCollectionInterval time.Duration
	BackupInterval         time.Duration
	DataFile               string
	Debug                  DebugConf
	Webserver              WebserverConf
	UVS232URL              string
}

type CurrentMeasurements struct {
	sync.Mutex
	solar.Values
}

type lastValues struct {
	state              solar.State
	stateDate          time.Time
	brinePumpState     solar.State
	brinePumpStateDate time.Time
	solarPumpState     solar.State
	solarPumpStateDate time.Time
}

type solarPumpRuntime struct {
	handler *solar.Measurements
	current *CurrentMeasurements
	last    lastValues
}

// Config holds the global configuration
var Config Configuration

// Measurements hold all measured heat pump values
var Measurements CurrentMeasurements
