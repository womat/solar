package main

import (
	"time"

	"github.com/womat/debug"
	"github.com/womat/tools"

	"solar/pkg/solar"
)

func (r *solarPumpRuntime) serveSave(f string, p time.Duration) {
	for range time.Tick(p) {
		_ = r.current.save(f)
	}
}

func (r *solarPumpRuntime) serveCalc(p time.Duration) {
	runtime := func(state solar.State, lastStateDate *time.Time, lastState *solar.State) (runTime float64) {
		if tools.In(state, solar.On, solar.HeatRecovery, solar.HeatingUpHotWater) {
			if !tools.In(*lastState, solar.On, solar.HeatRecovery, solar.HeatingUpHotWater) {
				*lastStateDate = time.Now()
			}
			runTime = time.Since(*lastStateDate).Hours()
			*lastStateDate = time.Now()
		}
		*lastState = state
		return
	}

	ticker := time.NewTicker(p)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		debug.DebugLog.Println("get data")

		v, err := r.handler.Read()
		if err != nil {
			debug.ErrorLog.Printf("get solar data: %v", err)
			continue
		}

		debug.DebugLog.Println("calc runtime")

		func() {
			r.current.Lock()
			defer r.current.Unlock()

			r.current.Timestamp = v.Timestamp
			r.current.CollectorTemp = v.CollectorTemp
			r.current.BoilerTemp = v.BoilerTemp
			r.current.State = v.State
			r.current.SolarFlow = v.SolarFlow
			r.current.SolarReturn = v.SolarReturn
			r.current.SolarPumpState = v.SolarPumpState
			r.current.BrineFlow = v.BrineFlow
			r.current.BrineReturn = v.BrineReturn
			r.current.BrinePumpState = v.BrinePumpState

			r.current.Runtime += runtime(r.current.State, &r.last.stateDate, &r.last.state)
			r.current.BrinePumpRuntime += runtime(r.current.BrinePumpState, &r.last.brinePumpStateDate, &r.last.brinePumpState)
			r.current.SolarPumpRuntime += runtime(r.current.SolarPumpState, &r.last.solarPumpStateDate, &r.last.solarPumpState)
		}()
	}
}
