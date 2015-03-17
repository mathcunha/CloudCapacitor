package capacitor

import (
	"fmt"
	"sort"
)

type Node struct {
	Height int
	Id     string
	Higher Configs
	Lower  Configs
	Equal  Configs
}

type DeploymentSpace struct {
	configs []*Configuration
}

type Configs []*Configuration
type byMem struct{ Configs }
type byCPU struct{ Configs }
type byPrice struct{ Configs }

func (s Configs) Len() int      { return len(s) }
func (s Configs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s byMem) Less(i, j int) bool   { return s.Configs[i].Mem() < s.Configs[j].Mem() }
func (s byCPU) Less(i, j int) bool   { return s.Configs[i].CPU() < s.Configs[j].CPU() }
func (s byPrice) Less(i, j int) bool { return s.Configs[i].Price() < s.Configs[j].Price() }

func NewDeploymentSpace(vms *[]VM, price float32, size int) (dspace DeploymentSpace) {
	conf := Configs{}
	atLeatOne := true
	for i := 1; i <= size && atLeatOne; i++ {
		atLeatOne = false
		for _, v := range *vms {
			c := Configuration{i, v}
			if c.Price() <= price {
				atLeatOne = true
				conf = append(conf, &c)
			}
		}
	}
	return DeploymentSpace{conf}
}

func (dspace *DeploymentSpace) CapacityBy(prop string) *map[string][]Node {
	switch prop {
	case "mem":
		sort.Sort(byMem{dspace.configs})
	case "cpu":
		sort.Sort(byCPU{dspace.configs})
	case "price":
		sort.Sort(byPrice{dspace.configs})
	}

	return nil
}

func (dspace DeploymentSpace) String() string {
	str := ""
	for _, v := range dspace.configs {
		str = fmt.Sprintf("%v%v\n", str, *v)
	}
	return str
}
