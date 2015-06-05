package capacitor

import (
	"fmt"
	"log"
	"reflect"
	"sort"
)

type Node struct {
	ID     string
	Level  int
	Higher Nodes
	Lower  Nodes
	Config *Configuration
}

type Nodes []*Node

type DeploymentSpace struct {
	configs *map[string]Configs
	cats    []string
}

type Configs []*Configuration
type byMem struct{ Configs }
type byCPU struct{ Configs }
type byPrice struct{ Configs }
type byStrict struct{ Configs }

type bySize struct{ Nodes }

func (s Nodes) Len() int      { return len(s) }
func (s Nodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s Configs) Len() int      { return len(s) }
func (s Configs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s byMem) Less(i, j int) bool    { return s.Configs[i].Mem() < s.Configs[j].Mem() }
func (s byCPU) Less(i, j int) bool    { return s.Configs[i].CPU() < s.Configs[j].CPU() }
func (s byPrice) Less(i, j int) bool  { return s.Configs[i].Price() < s.Configs[j].Price() }
func (s byStrict) Less(i, j int) bool { return s.Configs[i].Strict() < s.Configs[j].Strict() }

func (s bySize) Less(i, j int) bool { return s.Nodes[i].Config.Size < s.Nodes[j].Config.Size }

func NewDeploymentSpace(vms *[]VM, price float32, size int) (dspace DeploymentSpace) {
	mapa := make(map[string]Configs)
	categories := make(map[string]string)
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
				categories[c.Category] = c.Category
			}
		}
	}
	cats := make([]string, 0, 5)
	for cat, _ := range categories {
		cats = append(cats, cat)
	}
	sort.Strings(cats)
	return DeploymentSpace{&mapa, cats}
}

func (dspace *DeploymentSpace) Categories() []string {
	return dspace.cats
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
		//create each node
		for i, c := range configs {
			nodes[i] = new(Node)
			nodes[i].ID = fmt.Sprintf("%v_%v", c.Size, c.Name)
			nodes[i].Config = c
			if i == 0 {
				mapa[cat] = Nodes{}
			}
			mapa[cat] = append(mapa[cat], nodes[i])
		}
		//link nodes with their highers and lowers
		for _, out := range nodes {
			for _, in := range nodes {
				if out.Config.Name == in.Config.Name {
					if out.Config.Size+1 == in.Config.Size {
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
				if out.Config.Size == in.Config.Size {
					if out.Config.Strict()+1 == in.Config.Strict() {
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
		//set levels
		n := mapa[cat]
		root := n.NodeByID(fmt.Sprintf("%v_%v", 1, configs[0].Name))
		setLevels(root, 1)
	}
	return &mapa
}

func setLevels(n *Node, level int) {
	n.Level = level
	for _, l := range n.Lower {
		setLevels(l, level+1)
	}
}

func (nodes *Nodes) Equivalents(n *Node) Nodes {
	var equivalents Nodes
	for _, e := range *nodes {
		if n.ID != e.ID && n.Level == e.Level {
			equivalents = append(equivalents, e)
		}
	}
	sort.Sort(bySize{equivalents})
	return equivalents
}

func (dspace *DeploymentSpace) buildNodes(prop string) *map[string]Nodes {
	mapa := make(map[string]Nodes)

	for cat, configs := range *dspace.configs {
		nodes := make([]*Node, len(configs), len(configs))
		level := 0
		var v []reflect.Value

		//create each node
		for i, c := range configs {
			if i == 0 {
				level++
				v = reflect.ValueOf(c).MethodByName(prop).Call(nil)
			} else {
				value := reflect.ValueOf(c).MethodByName(prop).Call(nil)
				if !diffID(v[0], value[0]) {
					level++
					v = value
				}
			}
			nodes[i] = new(Node)
			nodes[i].ID = fmt.Sprintf("%v_%v", c.Size, c.Name)
			nodes[i].Config = c
			if i == 0 {
				mapa[cat] = Nodes{}
			}
			mapa[cat] = append(mapa[cat], nodes[i])
			nodes[i].Level = level
		}
		//link nodes with their highers and lowers
		for _, out := range nodes {
			for _, in := range nodes {
				if out.Level == in.Level+1 {
					//set higher
					if out.Higher == nil {
						out.Higher = Nodes{}
					}
					out.Higher = append(out.Higher, in)
				} else if out.Level == in.Level-1 {
					//set lower
					if out.Lower == nil {
						out.Lower = Nodes{}
					}
					out.Lower = append(out.Lower, in)
				}
			}
		}
	}
	return &mapa
}

//Not only equals, but with a diff between -0.001 and 0.001
func diffID(x reflect.Value, y reflect.Value) (equal bool) {
	switch y.Kind() {
	case reflect.Float32:
		diff := x.Float() - y.Float()
		return diff <= 0.001 && diff >= -0.001
	}

	return false
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
	str := fmt.Sprintf("{ id:%v, level:%v", n.ID, n.Level)
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
	return fmt.Sprintf("%v,config:%v}", str, n.Config)
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

func (nodes *Nodes) NodeByLevel(level int) (node *Node) {
	for _, node = range *nodes {
		if level == node.Level {
			return
		}
	}
	return nil
}

func NodesToDOT(mapNodes *map[string]Nodes) (str string) {
	colors := []string{"#FFCDD2", "#B2EBF2", "#C8E6C9", "#FFF9C4", "#E1BEE7", "#F0F4C3", "#D7CCC8", "#FFE0B2"}

	str = "\ngraph g {\n"

	edges := make(map[string]*[2]bool)

	for _, nodes := range *mapNodes {
		for _, node := range nodes {
			desc := fmt.Sprintf("%v (%v) [%v %v %v]", node.Config.Name, node.Config.Size, node.Config.CPU(), node.Config.Mem(), node.Config.Price())
			str = fmt.Sprintf("%v\"%v\" [label=\"%v\",shape=box,fillcolor=\"%v\",style=\"filled,rounded\"];\n", str, node.ID, desc, colors[node.Level%8])

			for _, lower := range node.Lower {
				key := fmt.Sprintf("\"%v\" -- \"%v\"", node.ID, lower.ID)
				if v, has := edges[key]; has {
					v[0] = true
				} else {
					edges[key] = &[2]bool{true, false}
				}
			}
			for _, higher := range node.Higher {
				key := fmt.Sprintf("\"%v\" -- \"%v\"", higher.ID, node.ID)
				if v, has := edges[key]; has {
					v[1] = true
				} else {
					edges[key] = &[2]bool{false, true}
				}
			}
		}
	}

	//the edges must exists from higher to lower and vice versa.
	//I could have used a bigraph, but the drawing sucks
	for key, v := range edges {
		if v[0] == v[1] {
			str = fmt.Sprintf("%v%v;\n", str, key)
		} else {
			log.Printf("ERROR: this key %v has must have two edges", key)
		}
	}

	str = fmt.Sprintf("%v}", str)
	return

}

func printTree(mapa *map[string]Nodes) string {
	str := ""
	m := *mapa
	for key, value := range m {
		str = fmt.Sprintf("%v\nKey=%v,Nodes=%v", str, key, value)
	}
	return str
}
