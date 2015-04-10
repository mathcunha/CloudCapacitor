package capacitor

import (
	"fmt"
	"github.com/mathcunha/CloudCapacitor/sync2"
	"strconv"
	"strings"
)

type Capacitor struct {
	Dspace DeploymentSpace
	Executor
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

func (c *Capacitor) HasMore(nodes *NodesInfo) bool {
	for _, node := range nodes.matrix {
		if node.When == -1 {
			return true
		}
	}
	return false
}

func (c *Capacitor) NodesLeft(nodes *NodesInfo) (count int) {
	for _, node := range nodes.matrix {
		if node.When == -1 {
			count++
		}
	}
	return
}

func (c *Capacitor) Exec(iNodes NodesInfo, slo float32, execs int, path string, wg *sync2.BlockWaitGroup, ch chan ExecInfo, it int, maxIts int) (int, string) {
	if it <= maxIts {
		endExecution := true
		for key, node := range iNodes.matrix {
			if !(node.When != -1) {
				endExecution = false
				cNodes := iNodes.Clone()
				nExecs := execs
				for _, conf := range node.Configs {
					nExecs = nExecs + 1

					result := c.Executor.Execute(*conf, node.WKL)

					cNodes.Mark(key, result.SLO <= slo, nExecs)

				}
				_, err := wg.Add(1)
				nPath := fmt.Sprintf("%v%v->", path, key)
				if err == nil {
					go func() {
						defer wg.Done()
						c.Exec(*cNodes, slo, nExecs, nPath, wg, ch, it+1, maxIts)
					}()
				} else {
					c.Exec(*cNodes, slo, nExecs, nPath, wg, ch, it+1, maxIts)
				}
			}
		}
		if endExecution {
			//All executions!
			ch <- ExecInfo{execs, path}
			return execs, path
		} else {
			return -1, "NOTHING"
		}
	} else {
		return -1, "MAX ITs REACHED"
	}
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
		matrix[key].Reject = true
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

func (node NodeInfo) String() (str string) {
	return fmt.Sprintf("{when:%v, exec:%v, candidate:%v, reject:%v}", node.When, node.Exec, node.Candidate, node.Reject)
}

func (nodes NodesInfo) String() (str string) {
	str = fmt.Sprintf("{wkl:%v, heigh:%v, nodes:[\n\t", nodes.lenWKL, nodes.lenNodes)
	for key, n := range nodes.matrix {
		str = fmt.Sprintf("%v[%v]%v\n\t", str, key, n)
	}
	str = fmt.Sprintf("%v]}", str)
	return
}
