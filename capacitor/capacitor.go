package capacitor

import (
	"fmt"
	"strconv"
	"strings"
)

type Capacitor struct {
	dspace DeploymentSpace
	exec   Executor
}

type NodesInfo struct {
	matrix   map[string]NodeInfo
	lenWKL   int
	lenNodes int
}

type NodeInfo struct {
	Node
	WKL       string
	Exec      bool
	Reject    bool
	Candidate bool
	When      int
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

func (c *Capacitor) MinExec(mode string, slo float32, wkls []string) {
	mapa := c.dspace.CapacityBy(mode)
	for _, nodes := range *mapa {
		c.Exec(buildMatrix(wkls, nodes), slo, 0, "")
	}
}

func buildMatrix(wkls []string, nodes Nodes) (matrix NodesInfo) {
	iNodes := NodesInfo{make(map[string]NodeInfo), -1, -1}
	max := -1
	for _, node := range nodes {
		for i, wkl := range wkls {
			iNodes.matrix[fmt.Sprintf("%v_%v", node.Height, i)] = NodeInfo{*node, wkl, false, false, false, -1}
		}
		max = node.Height
	}
	iNodes.lenWKL = len(wkls)
	iNodes.lenNodes = max
	return iNodes
}

func (c *Capacitor) Exec(iNodes NodesInfo, slo float32, execs int, path string) (int, string) {
	for key, node := range iNodes.matrix {
		noneAvailable := true
		if !(node.When != -1) {
			noneAvailable = false
			for _, conf := range node.Configs {
				execs = execs + 1
				node.When = execs

				result := c.exec.Execute(*conf, node.WKL)
				clone := iNodes.Clone()
				pClone := &clone
				pINode := &iNodes

				pINode.Mark(key, result.SLO <= slo, execs)
				fmt.Printf("%v %v\n", key, pClone.Equal(pINode))

				//fmt.Printf("%v x %v ? %v \n", *conf, node.WKL, result.SLO <= slo)
				//Faço as marcações
				//Verifica se tem mais node para chamar
				//Exec(slo, wkls, nodes, execs, path)
			}
		}
		if noneAvailable {
			//All executions!
			fmt.Printf("%v, %v", execs, path)
			return execs, path
		}
	}
	return -1, ""
}

func (pNodes *NodesInfo) Mark(key string, metslo bool, exec int) {
	s := strings.Split(key, "_")
	cHeight, _ := strconv.ParseInt(s[0], 0, 64)
	cWKL, _ := strconv.ParseInt(s[1], 0, 64)
	//fmt.Printf("%v_%v \n", cHeight, cWKL)
	iNodes := *pNodes
	matrix := iNodes.matrix

	if metslo {
		for height := int64(1); height < cHeight; height++ {
			for i := cWKL; i >= 0; i-- {
				nodeInfo := matrix[fmt.Sprintf("%v_%v", height, i)]
				if nodeInfo.When != -1 {
					nodeInfo.Candidate = true
					nodeInfo.When = exec
				}
			}
		}
	} else {
		for height := cHeight - 1; height >= 1; height-- {
			for i := int64(0); i <= cWKL; i++ {
				nodeInfo := matrix[fmt.Sprintf("%v_%v", height, i)]
				if nodeInfo.When != -1 {
					nodeInfo.Reject = true
					nodeInfo.When = exec
				}
			}
		}
	}
}
func (node *NodeInfo) Equal(other *NodeInfo) bool {
	return node.Height == other.Height && node.ID == other.ID && node.Exec == other.Exec && node.Reject == other.Reject && node.Candidate == other.Candidate && node.When == other.When
}

func (iNode *NodesInfo) Equal(other *NodesInfo) bool {
	clone := *other
	matrix := *iNode
	for key, n := range matrix.matrix {
		oNode := clone.matrix[key]
		node := &n
		if !node.Equal(&oNode) {
			return false
		}
	}
	return true
}

func (matrix NodesInfo) Clone() (clone NodesInfo) {
	clone = matrix
	return clone
}
