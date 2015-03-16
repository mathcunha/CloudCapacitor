package capacitor

import (
	"fmt"
)

type DeploymentSpace struct {
	configs []Configuration
}

func NewDeploymentSpace(vms *[]VM, price float32, size int) (dspace DeploymentSpace) {
	conf := []Configuration{}
	atLeatOne := true
	for i := 1; i <= size && atLeatOne; i++ {
		atLeatOne = false
		for _, v := range *vms {
			c := Configuration{i, v}
			if c.Price() <= price {
				atLeatOne = true
				conf = append(conf, c)
			}
		}
	}
	return DeploymentSpace{conf}
}

func (dspace *DeploymentSpace) CapacityBy(prop string) *DeploymentSpace {
	return dspace
}

func (dspace *DeploymentSpace) String() string {
	str := ""
	for _, v := range dspace.configs {
		str = fmt.Sprintf("%v%v\n", str, v)
	}
	return str
}
