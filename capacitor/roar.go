package capacitor

import (
	"fmt"
	"strconv"
)

//Vitual Machine types, workload list, throughtput expected, executor
func findPeakByType(vms *[]VM, wkls []string, te float32, e Executor) (map[string]float32, error) {
	throughtputs := make(map[string]float32)
	for _, vm := range *vms {
		fmt.Printf("ROAR VM %v\n", vm)
		peak := float32(0.0)
		for _, wkl := range wkls {
			fmt.Printf("ROAR WKL %v\n", wkl)
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
	return throughtputs, nil
}
