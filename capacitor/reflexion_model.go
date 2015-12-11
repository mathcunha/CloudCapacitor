package capacitor

import (
	"strings"
)

func VerifyReflexionModel(configs Configs, mapNodes *map[string]Nodes) (convergence, absence, divergence int) {
	for _, c := range configs {
		node := getNodeByConf(c, mapNodes)
		for _, lNode := range node.Lower {
			//maxslo by workload
			//fmt.Printf("[VerifyReflexionModel] %s -> %s\n", node.Config, lNode.Config)
			for i := 0; i < len(node.Config.MaxSLO()); i++ {
				result := strings.Compare(node.Config.MaxSLO()[i:i+1], lNode.Config.MaxSLO()[i:i+1])
				//fmt.Printf("[VerifyReflexionModel] MaxSLOID h:%s l:%s compare:%d\n", node.Config.MaxSLO()[i:i+1], lNode.Config.MaxSLO()[i:i+1], result)
				switch result {
				case 1:
					convergence += 2
				case 0:
					convergence += 2
				case -1:
					divergence += 2
				}
			}
		}
		//same level possible absence
		for k := 0; k < len(node.Lower); k++ {
			for j := k + 1; j < len(node.Lower); j++ {
				for i := 0; i < len(node.Config.MaxSLO()); i++ {
					result := strings.Compare(node.Lower[k].Config.MaxSLO()[i:i+1], node.Lower[j].Config.MaxSLO()[i:i+1])
					if result == 0 {
						absence += 2
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
		srcNode := getNodeByConf(c, src)
		destNode := getNodeByConf(c, dest)
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

func getNodeByConf(conf *Configuration, ds *map[string]Nodes) (node *Node) {
	node = nil
	for _, nodes := range *ds {
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
