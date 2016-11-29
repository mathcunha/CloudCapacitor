package capacitor

import (
	"log"
	"strconv"
)

type ROARExecutor struct {
	Executor
	throughtputs map[string]float32
}

func (e *ROARExecutor) Execute(config Configuration, wkl string) (r Result) {
	if config.Size == 1 {
		//because it was already executed at NewROARExecutor method. //TODO store raw execution results
		return e.Executor.Execute(config, wkl)
	}
	if iWKL, err := strconv.ParseInt(wkl, 10, 32); err == nil {
		r = Result{Config: config, Performance: Performance{SLO: float32(iWKL) / e.throughtputs[config.VM.Name] * float32(config.Size)}}
		r.NotExected = true
	} else {
		log.Printf("ROARExecutor.Execute, error parsing %v to int", wkl)
		r = Result{}
	}
	return
}

//Vitual Machine types, workload list, throughtput expected, executor
func NewROARExecutor(vms *[]VM, wkls []string, e Executor) (*ROARExecutor, error) {
	throughtputs := make(map[string]float32)
	for _, vm := range *vms {
		peak := float32(0.0)
		for _, wkl := range wkls {
			result := e.Execute(Configuration{Size: 1, VM: vm}, wkl)
			if iWKL, err := strconv.ParseInt(wkl, 10, 32); err != nil {
				return nil, err
			} else {
				throughtput := float32(iWKL) / result.SLO
				if peak < throughtput {
					peak = throughtput
				} else {
					break
				}
			}
		}
		throughtputs[vm.Name] = peak
	}
	return &ROARExecutor{e, throughtputs}, nil
}
