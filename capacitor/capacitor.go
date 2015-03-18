package capacitor

import (
	"fmt"
)

type Capacitor struct {
	dspace DeploymentSpace
	exec   Executor
}

//Brutal force heuristic
func (c *Capacitor) BF(mode string, slo float32, wkls []string) {
	mapa := c.dspace.CapacityBy(mode)
	for _, nodes := range *mapa {
		for _, node := range nodes {
			for _, conf := range node.Configs {
				for _, wkl := range wkls {
					result := c.exec.Execute(*conf, wkl)
					fmt.Printf("%v x %v ? %v \n", *conf, wkl, result.SLO <= slo)
				}
			}
		}
	}
}
