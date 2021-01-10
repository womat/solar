package solar

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/womat/debug"
)

const httpRequestTimeout = 10 * time.Second

const (
	On                State = "on"
	Off               State = "off"
	HeatingUpHotWater State = "heating up hot water"
	HeatRecovery      State = "heat recovery"
)

type State string

type Values struct {
	Timestamp              time.Time
	CollectorTemp          float64
	BoilerTemp             float64
	State                  State
	Runtime                float64
	SolarFlow, SolarReturn float64
	SolarPumpState         State
	SolarPumpRuntime       float64
	BrineFlow, BrineReturn float64
	BrinePumpState         State
	BrinePumpRuntime       float64
}

type Measurements struct {
	uvs232URL string
}

type uvs232URLBody struct {
	Timestamp time.Time `json:"Timestamp"`
	Runtime   float64   `json:"Runtime"`
	Measurand struct {
		Temperature1, Temperature2, Temperature3, Temperature4 float64
		Out1, Out2                                             bool
		RotationSpeed                                          float64
	} `json:"Data"`
}

func New() *Measurements {
	return &Measurements{}
}

func (m *Measurements) SetUVS232URL(url string) {
	m.uvs232URL = url
}

func (m *Measurements) Read() (v Values, err error) {
	start := time.Now()

	if v, err = m.readUVS232(); err != nil {
		return
	}

	debug.DebugLog.Printf("runtime to request data: %vs", time.Since(start).Seconds())

	switch {
	case v.SolarPumpState == On && v.BrinePumpState == On:
		v.State = HeatRecovery
	case v.SolarPumpState == On && v.BrinePumpState == Off:
		v.State = HeatingUpHotWater
	default:
		v.State = Off
	}

	return
}

func (m *Measurements) readUVS232() (v Values, err error) {
	var r uvs232URLBody

	if err = read(m.uvs232URL, &r); err != nil {
		return
	}

	v.Timestamp = r.Timestamp
	v.CollectorTemp = r.Measurand.Temperature1
	v.BoilerTemp = r.Measurand.Temperature2
	v.BrineFlow = r.Measurand.Temperature3
	v.BrineReturn = r.Measurand.Temperature4

	if r.Measurand.Out1 {
		v.SolarPumpState = On
	} else {
		v.SolarPumpState = Off
	}
	if r.Measurand.Out2 {
		v.BrinePumpState = On
	} else {
		v.BrinePumpState = Off
	}
	return
}

func read(url string, data interface{}) (err error) {
	done := make(chan bool, 1)
	go func() {
		// ensures that data is sent to the channel when the function is terminated
		defer func() {
			select {
			case done <- true:
			default:
			}
			close(done)
		}()

		debug.TraceLog.Printf("performing http get: %v\n", url)

		var res *http.Response
		if res, err = http.Get(url); err != nil {
			return
		}

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		_ = res.Body.Close()

		if err = json.Unmarshal(bodyBytes, data); err != nil {
			return
		}
	}()

	// wait for API Data
	select {
	case <-done:
	case <-time.After(httpRequestTimeout):
		err = errors.New("timeout during receive data")
	}

	if err != nil {
		debug.ErrorLog.Println(err)
		return
	}
	return
}
