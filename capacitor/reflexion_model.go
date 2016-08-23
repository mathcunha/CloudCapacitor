package capacitor

import (
	"fmt"
	"strings"
)

const (
	Bigger = iota
	Smaller
	Equal
	Empty
)

func NameRelation(rel int) string {
	switch rel {
	case Bigger:
		return "<--"
	case Smaller:
		return "-->"
	case Equal:
		return "<->"
	case Empty:
		return " O "
	}
	return ""
}

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

func WhatRelation(n1, n2 *Node, key1, key2 string, ds *map[string]Nodes, equi bool) (relation int) {
	relation = Empty
	if key1 == key2 {
		nodes := (*ds)[key1]
		if n1.Level < n2.Level {
			levelNodes := n1.Lower
			if equi {
				levelNodes = nodes.FromLevel(n1.Level + 1)
			}
			if hasPathUntilNode(n2, levelNodes, &nodes, equi) {
				return Smaller
			} else {
				return Empty
			}
		} else if n1.Level > n2.Level {
			levelNodes := n2.Lower
			if equi {
				levelNodes = nodes.FromLevel(n2.Level + 1)
			}
			if hasPathUntilNode(n1, levelNodes, &nodes, equi) {
				return Bigger
			} else {
				return Empty
			}
		} else {
			if equi {
				return Equal
			} else {
				return Empty
			}
		}

	}
	return
}

func hasPathUntilNode(target *Node, levelNodes Nodes, nodes *Nodes, equi bool) (has bool) {
	if levelNodes == nil || len(levelNodes) == 0 {
		has = false
		return
	}
	if target.Level == levelNodes[0].Level {
		for _, node := range levelNodes {
			if confEqual(node.Config, target.Config) {
				has = true
				return
			}
		}
		has = false
		return
	}
	if equi {
		has = hasPathUntilNode(target, nodes.FromLevel(levelNodes[0].Level+1), nodes, equi)
		return
	}
	for _, node := range levelNodes {
		has = hasPathUntilNode(target, node.Lower, nodes, equi)
		if has {
			break
		}
	}
	return
}

func CalcCapacityAccuracy (convergence, absence, divergence int) (result float32){
	result = float32 (convergence) / float32 (convergence + divergence + (2 * absence))
	return result
}

func VerifyReflexionModel3(c Configs, mapNodes *map[string]Nodes, equi bool) (convergence, absence, divergence int) {
	hasArcN1_N2 := func(r int) bool {
		switch r {
		case Equal:
			return true
		case Bigger:
			return true
		}
		return false

	}
	hasArcN2_N1 := func(r int) bool {
		switch r {
		case Equal:
			return true
		case Smaller:
			return true
		}
		return false

	}
	calcModel := func(hasArcModel, hasArcSource bool) (int, int, int) {
		if hasArcSource && hasArcModel && true {
			return 1, 0, 0 //convergence
		}
		if hasArcSource && !hasArcModel {
			return 0, 0, 1 //divergence
		}
		if hasArcModel && !hasArcSource {
			return 0, 1, 0 //absence
		}
		return 0, 0, 0
	}
	for k := 0; k < len(c); k++ {
		for j := k + 1; j < len(c); j++ {
			n1, key1 := getNodeByConf(c[k], mapNodes)
			n2, key2 := getNodeByConf(c[j], mapNodes)
			dsRelation := WhatRelation(n1, n2, key1, key2, mapNodes, equi)
			realRelation := Bigger
			if strings.Compare(n1.Config.MaxSLO(), n2.Config.MaxSLO()) == 0 {
				realRelation = Equal
			} else if isConvergence(n1, n2) {
				realRelation = Smaller
			}
			c, a, d := calcModel(hasArcN1_N2(realRelation), hasArcN1_N2(dsRelation))
			convergence += c
			absence += a
			divergence += d
			c, a, d = calcModel(hasArcN2_N1(realRelation), hasArcN2_N1(dsRelation))
			convergence += c
			absence += a
			divergence += d
		}
	}
	return
}

func VerifyReflexionModel2(c Configs, mapNodes *map[string]Nodes, equi bool) (convergence, absence, divergence int) {
	for k := 0; k < len(c); k++ {
		for j := k + 1; j < len(c); j++ {
			n1, key1 := getNodeByConf(c[k], mapNodes)
			n2, key2 := getNodeByConf(c[j], mapNodes)
			dsRelation := WhatRelation(n1, n2, key1, key2, mapNodes, equi)
			realRelation := Bigger
			if strings.Compare(n1.Config.MaxSLO(), n2.Config.MaxSLO()) == 0 {
				realRelation = Equal
			} else if isConvergence(n1, n2) {
				realRelation = Smaller
			}
			if dsRelation == Empty {
				switch realRelation {
				case Smaller:
					absence++
				case Equal:
					absence += 2
				case Bigger:
					absence++
				}
			} else if dsRelation == Smaller {
				switch realRelation {
				case Smaller:
					convergence++
				case Bigger:
					divergence++
					absence++
				case Equal:
					convergence++
					absence++
				}
			} else if dsRelation == Bigger {
				switch realRelation {
				case Bigger:
					convergence++
				case Smaller:
					divergence++
					absence++
				case Equal:
					convergence++
					absence++
				}
			} else if dsRelation == Equal {
				switch realRelation {
				case Bigger:
					convergence++
					divergence++
				case Smaller:
					convergence++
					divergence++
				case Equal:
					convergence += 2
				}
			}
			//fmt.Printf("%s(%d)\t%s\t%s(%d)\tReal:%s\t(c:%d,a:%d,d:%d)\n", n1.Config.Name, n1.Config.Size, NameRelation(dsRelation), n2.Config.Name, n2.Config.Size, NameRelation(realRelation), convergence, absence, divergence)
		}
	}
	return
}

func VerifyReflexionModel(configs Configs, mapNodes *map[string]Nodes, equi bool) (convergence, absence, divergence int) {
	for _, c := range configs {
		node, key := getNodeByConf(c, mapNodes)
		nodes := (*mapNodes)[key]
		for _, lNode := range node.Lower {
			//maxslo by workload
			if isConvergence(node, lNode) {
				convergence++
			} else {
				divergence++
			}

		}
		if node.Lower != nil && len(node.Lower) > 0 {
			levelNodes := nodes.FromLevel(node.Lower[0].Level)
			for _, levelNode := range levelNodes {
				lower := false
				for _, lNode := range node.Lower {
					if confEqual(levelNode.Config, lNode.Config) {
						lower = true
						break
					}
				}
				if strings.Compare(node.Config.MaxSLO(), levelNode.Config.MaxSLO()) == 0 {
					if !lower {
						absence += 2
					} else {
						absence++
					}
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
				levelNodes := nodes.FromLevel(n.Level)
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

func getNodeByConf(conf *Configuration, ds *map[string]Nodes) (node *Node, key string) {
	node = nil
	for key, nodes := range *ds {
		for _, n := range nodes {
			if confEqual(conf, n.Config) {
				return n, key
			}
		}
	}
	return
}

func confEqual(src, dest *Configuration) bool {
	return (src.Size == dest.Size && src.VM.Name == dest.VM.Name)
}
