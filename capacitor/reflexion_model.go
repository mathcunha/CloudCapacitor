package capacitor

import (
	"fmt"
	"strings"
)

func isConvergence(node, lNode *Node) bool {
	//fmt.Printf("[VerifyReflexionModel] %s -> %s\n", node.Config, lNode.Config)
	for i := 0; i < len(node.Config.MaxSLO()); i++ {
		result := strings.Compare(node.Config.MaxSLO()[i:i+1], lNode.Config.MaxSLO()[i:i+1])
		//fmt.Printf("[VerifyReflexionModel] MaxSLOID h:%s l:%s compare:%d\n", node.Config.MaxSLO()[i:i+1], lNode.Config.MaxSLO()[i:i+1], result)
		switch result {
		case -1:
			return false
		}
	}
	return true
}

func printAbscence(a, b *Node, both bool) {
	vertex := "->"
	if both {
		vertex = "<->"
	}
	fmt.Printf("%s(%d)\t%s\t%s(%d)\n", a.Config.Name, a.Config.Size, vertex, b.Config.Name, b.Config.Size)
}

func VerifyReflexionModel(configs Configs, mapNodes *map[string]Nodes, equi bool) (convergence, absence, divergence int) {
	for _, c := range configs {
		node, nodes := getNodeByConf(c, mapNodes)
		for _, lNode := range node.Lower {
			//maxslo by workload
			if isConvergence(node, lNode) {
				convergence++
			} else {
				divergence++
			}

		}
		if node.Lower != nil && len(node.Lower) > 0 {
			levelNodes := nodes.FromLevel(node.Lower[0])
			for _, levelNode := range levelNodes {
				lower := false
				for _, lNode := range node.Lower {
					if confEqual(levelNode.Config, lNode.Config) {
						lower = true
						break
					}
				}
				if strings.Compare(node.Config.MaxSLO(), levelNode.Config.MaxSLO()) == 0 {
					absence++
				} else if !lower {
					if isConvergence(node, levelNode) {
						absence++
					}
				}
			}
		}
	}
	//same level possible absence
	for _, nodes := range *mapNodes {
		isLevelVerified := make([]bool, len(configs), len(configs))
		for _, n := range nodes {
			if !isLevelVerified[n.Level] {
				isLevelVerified[n.Level] = true
				levelNodes := nodes.FromLevel(n)
				for k := 0; k < len(levelNodes); k++ {
					for j := k + 1; j < len(levelNodes); j++ {
						result := strings.Compare(levelNodes[k].Config.MaxSLO(), levelNodes[j].Config.MaxSLO())
						if result == 0 {
							if equi {
								convergence += 2
							} else {
								absence += 2
							}
						} else {
							if equi {
								divergence++
								convergence++
							} else {
								absence++
							}
						}
					}
				}
			}
		}
	}
	//fmt.Printf("[VerifyReflexionModel] convergence:%d, absence:%d, divergence:%d\n", convergence, absence, divergence)
	return
}

func DiffDS(configs Configs, src, dest *map[string]Nodes) (convergence, absence, divergence int) {
	for _, c := range configs {
		srcNode, _ := getNodeByConf(c, src)
		destNode, _ := getNodeByConf(c, dest)
		c, a, d := nodeDiff(srcNode, destNode)
		convergence += c
		absence += a
		divergence += d
	}
	return
}

func nodeDiff(srcNode, destNode *Node) (convergence, absence, divergence int) {
	//fmt.Printf("\tsrcNode:%s\ndestNode:%s\n\n", srcNode, destNode)
	c, a, d := nodesArrayDiff(srcNode.Higher, destNode.Higher)
	convergence += c
	absence += a
	divergence += d
	c, a, d = nodesArrayDiff(srcNode.Lower, destNode.Lower)
	convergence += c
	absence += a
	divergence += d
	//fmt.Printf("nodeDiff:\tconvergence:%d, absence:%d, divergence:%d\n", convergence, absence, divergence)
	return
}

func nodesArrayDiff(src, dest Nodes) (convergence, absence, divergence int) {
	if src != nil || dest != nil {
		for _, srcNode := range src {
			found := false
			for _, destNode := range dest {
				found = found || confEqual(srcNode.Config, destNode.Config)
			}
			if found {
				convergence++
			} else {
				absence++
			}
		}

		for _, destNode := range dest {
			found := false
			for _, srcNode := range src {
				found = found || confEqual(srcNode.Config, destNode.Config)
			}
			if !found {
				divergence++
			}
		}
	}
	return
}

func getNodeByConf(conf *Configuration, ds *map[string]Nodes) (node *Node, nodes Nodes) {
	node = nil
	for _, nodes = range *ds {
		for _, n := range nodes {
			if confEqual(conf, n.Config) {
				node = n
				return
			}
		}
	}
	return
}

func confEqual(src, dest *Configuration) bool {
	return (src.Size == dest.Size && src.VM.Name == dest.VM.Name)
}
