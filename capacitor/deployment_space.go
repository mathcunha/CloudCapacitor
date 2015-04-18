package capacitor

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

type Node struct {
	ID     string
	Height int
	Higher Nodes
	Lower  Nodes
	Configs
}

type Nodes []*Node

type DeploymentSpace struct {
	configs *map[string]Configs
}

type Configs []*Configuration
type byMem struct{ Configs }
type byCPU struct{ Configs }
type byPrice struct{ Configs }
type byStrict struct{ Configs }

func (s Configs) Len() int      { return len(s) }
func (s Configs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s byMem) Less(i, j int) bool    { return s.Configs[i].Mem() < s.Configs[j].Mem() }
func (s byCPU) Less(i, j int) bool    { return s.Configs[i].CPU() < s.Configs[j].CPU() }
func (s byPrice) Less(i, j int) bool  { return s.Configs[i].Price() < s.Configs[j].Price() }
func (s byStrict) Less(i, j int) bool { return s.Configs[i].Strict() < s.Configs[j].Strict() }

func NewDeploymentSpace(vms *[]VM, price float32, size int) (dspace DeploymentSpace) {
	mapa := make(map[string]Configs)
	atLeatOne := true
	for i := 1; i <= size && atLeatOne; i++ {
		atLeatOne = false
		for _, v := range *vms {
			c := Configuration{i, v}
			if c.Price() <= price {
				atLeatOne = true
				conf := mapa[c.Category]
				conf = append(conf, &c)
				mapa[c.Category] = conf
			}
		}
	}
	return DeploymentSpace{&mapa}
}

func (dspace *DeploymentSpace) CapacityBy(prop string) (list *map[string]Nodes) {
	for _, v := range *dspace.configs {
		switch prop {

		case "Mem":
			sort.Sort(byMem{v})
		case "CPU":
			sort.Sort(byCPU{v})
		case "Price":
			sort.Sort(byPrice{v})
		case "Strict":
			sort.Sort(byStrict{v})
			return dspace.buildNodesStrict(prop)
		}

	}

	return dspace.buildNodes(prop)
}

func (dspace *DeploymentSpace) buildNodesStrict(prop string) *map[string]Nodes {
	mapa := make(map[string]Nodes)

	for cat, configs := range *dspace.configs {
		nodes := make([]*Node, len(configs), len(configs))
		for i, c := range configs {
			nodes[i] = new(Node)
			nodes[i].ID = fmt.Sprintf("%v_%v", c.Size, c.Name)
			nodes[i].Configs = Configs{c}
			if i == 0 {
				mapa[cat] = Nodes{}
			}
			mapa[cat] = append(mapa[cat], nodes[i])
		}
		for _, out := range nodes {
			for _, in := range nodes {
				if out.Configs[0].Name == in.Configs[0].Name {
					if out.Configs[0].Size+1 == in.Configs[0].Size {
						if out.Lower == nil {
							out.Lower = Nodes{}
						}
						out.Lower = append(out.Lower, in)
						if in.Higher == nil {
							in.Higher = Nodes{}
						}
						in.Higher = append(in.Higher, out)
					}
				}
				if out.Configs[0].Size == in.Configs[0].Size {
					if out.Configs[0].Strict()*2 == in.Configs[0].Strict() {
						if out.Lower == nil {
							out.Lower = Nodes{}
						}
						out.Lower = append(out.Lower, in)
						if in.Higher == nil {
							in.Higher = Nodes{}
						}
						in.Higher = append(in.Higher, out)
					}
				}
			}
		}
	}
	return &mapa
}

func (dspace *DeploymentSpace) buildNodes(prop string) *map[string]Nodes {
	mapa := make(map[string]Nodes)
	for cat, configs := range *dspace.configs {
		v := reflect.ValueOf(configs[0]).MethodByName(prop).Call(nil)
		node := new(Node)
		equal, ID := equalID(v[0], v[0])
		node.ID = ID
		node.Height = 0
		for _, c := range configs {
			equal, ID = equalID(v[0], reflect.ValueOf(c).MethodByName(prop).Call(nil)[0])
			if equal {
				node.Configs = append(node.Configs, c)
			} else {
				updateMap(&mapa, node, cat)
				node = new(Node)
				node.ID = ID
				node.Configs = append(node.Configs, c)
				v = reflect.ValueOf(c).MethodByName(prop).Call(nil)
			}
		}

		if node != nil {
			updateMap(&mapa, node, cat)
		}
	}

	return &mapa
}

func updateMap(mapa *map[string]Nodes, node *Node, cat string) {
	m := *mapa
	n, has := m[cat]
	if has {
		max := len(n)
		nodes := append(n, node)

		nodes[max].Higher = Nodes{nodes[max-1]}
		nodes[max-1].Lower = Nodes{nodes[max]}
		nodes[max].Height = max + 1

		m[cat] = nodes
	} else {
		node.Higher = nil
		node.Height = 1

		m[cat] = Nodes{node}
	}
	node = nil
}

func equalID(x reflect.Value, y reflect.Value) (equal bool, id string) {
	switch y.Kind() {
	case reflect.Float32:
		return x.Float() == y.Float(), strconv.FormatFloat(y.Float(), 'f', 2, 32)
	}

	return false, ""
}

func (dspace DeploymentSpace) String() string {
	str := ""
	for _, confs := range *dspace.configs {
		for _, v := range confs {
			str = fmt.Sprintf("%v%v\n", str, *v)
		}
	}
	return str
}

func (n Node) String() string {
	str := fmt.Sprintf("{ id:%v, height:%v", n.ID, n.Height)
	if n.Higher != nil {
		str = fmt.Sprintf("%v, higher:", str)
		for i, h := range n.Higher {
			str = fmt.Sprintf("%v%v", str, h.ID)
			if i+1 != len(n.Higher) {
				str = fmt.Sprintf("%v,", str)
			}
		}
	} else {
		str = fmt.Sprintf("%v, root:true", str)
	}
	if n.Lower != nil {
		str = fmt.Sprintf("%v, lower:", str)
		for i, l := range n.Lower {
			str = fmt.Sprintf("%v%v", str, l.ID)
			if i+1 != len(n.Lower) {
				str = fmt.Sprintf("%v,", str)
			}
		}
	} else {
		str = fmt.Sprintf("%v, leaf:true", str)
	}
	return fmt.Sprintf("%v,configs:%v}", str, n.Configs)
}

func (nodes Nodes) String() string {
	str := ""
	for _, v := range nodes {
		str = fmt.Sprintf("%v,%v\n", str, v)
	}
	return str
}

func (nodes *Nodes) NodeByID(ID string) (node *Node) {
	for _, node = range *nodes {
		if ID == node.ID {
			return
		}
	}
	return nil
}

func printTree(mapa *map[string]Nodes) string {
	str := ""
	m := *mapa
	for key, value := range m {
		str = fmt.Sprintf("%v\nKey=%v,Nodes=%v", str, key, value)
	}
	return str
}
