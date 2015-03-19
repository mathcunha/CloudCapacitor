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
	matrix   map[string]*NodeInfo
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
	//fmt.Println(mapa)
	for _, nodes := range *mapa {
		c.ExecCategory(wkls, nodes, slo)
	}
}

func (c *Capacitor) ExecCategory(wkls []string, nodes Nodes, slo float32) {
	matrix := buildMatrix(wkls, nodes)
	//fmt.Println(matrix)
	c.Exec(matrix, slo, 0, "")
	//fmt.Printf("%v - %v\n", exec, path)

}

func buildMatrix(wkls []string, nodes Nodes) (matrix NodesInfo) {
	iNodes := NodesInfo{make(map[string]*NodeInfo), -1, -1}
	max := -1
	for _, node := range nodes {
		for i, wkl := range wkls {
			n := new(NodeInfo)
			n.Node = *node
			n.WKL = wkl
			n.Exec = false
			n.Reject = false
			n.Candidate = false
			n.When = -1
			iNodes.matrix[fmt.Sprintf("%v_%v", node.Height, i)] = n
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
			cNodes := iNodes.Clone()
			for _, conf := range node.Configs {
				execs = execs + 1

				result := c.exec.Execute(*conf, node.WKL)

				cNodes.Mark(key, result.SLO <= slo, execs)

			}

			c.Exec(*cNodes, slo, execs, fmt.Sprintf("%v%v,", path, key))
		}
		if noneAvailable {
			//All executions!
			fmt.Printf("#execs:%v, path:%v \n", execs, path)
			return execs, path
		}
	}
	return -1, "WARNING - NO NODE"
}

func (pNodes *NodesInfo) Mark(key string, metslo bool, exec int) {
	s := strings.Split(key, "_")
	cHeight, _ := strconv.ParseInt(s[0], 0, 64)
	cWKL, _ := strconv.ParseInt(s[1], 0, 64)
	//fmt.Printf("INI MARK\n")
	//fmt.Printf("%v ? %v\n", key, metslo)
	iNodes := *pNodes
	matrix := iNodes.matrix

	matrix[key].When = exec
	matrix[key].Exec = true

	if metslo {
		matrix[key].Candidate = true
		for height := cHeight; height <= int64(pNodes.lenNodes); height++ {
			for i := cWKL; i >= 0; i-- {
				key := fmt.Sprintf("%v_%v", height, i)
				nodeInfo := matrix[key]
				if nodeInfo.When == -1 {
					nodeInfo.Candidate = true
					nodeInfo.When = exec
					//fmt.Println("\tmarcando ", key, " candidato ")
				}
			}
		}
	} else {
		matrix[key].Candidate = false
		for height := cHeight; height >= 1; height-- {
			for i := cWKL; i < int64(pNodes.lenWKL); i++ {
				key := fmt.Sprintf("%v_%v", height, i)
				nodeInfo := matrix[key]
				if nodeInfo.When == -1 {
					nodeInfo.Reject = true
					nodeInfo.When = exec
					//fmt.Println("\tmarcando ", key, " rejeitado ")
				}
			}
		}
	}
	//fmt.Printf("FIM MARK\n")
}
func (node *NodeInfo) Equal(other *NodeInfo) bool {
	return node.Height == other.Height && node.ID == other.ID && node.Exec == other.Exec && node.Reject == other.Reject && node.Candidate == other.Candidate && node.When == other.When
}

func (iNode *NodesInfo) Equal(other *NodesInfo) bool {
	clone := *other
	matrix := *iNode
	for key, n := range matrix.matrix {
		if !n.Equal(clone.matrix[key]) {
			return false
		}
	}
	return true
}

func (matrix NodesInfo) Clone() (clone *NodesInfo) {
	mapa := make(map[string]*NodeInfo)
	for key, node := range matrix.matrix {
		n := new(NodeInfo)
		n.Node = node.Node
		n.WKL = node.WKL
		n.Exec = node.Exec
		n.Reject = node.Reject
		n.Candidate = node.Candidate
		n.When = node.When

		mapa[key] = n
	}
	pClone := new(NodesInfo)
	pClone.matrix = mapa
	pClone.lenWKL = matrix.lenWKL
	pClone.lenNodes = matrix.lenNodes
	return pClone
}
