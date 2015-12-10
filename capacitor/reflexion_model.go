package capacitor

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
