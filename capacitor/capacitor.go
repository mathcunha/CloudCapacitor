package capacitor

import (
	"fmt"
	"strconv"
	"strings"
)

type Capacitor struct {
	Dspace DeploymentSpace
	Executor
}

type NodesInfo struct {
	Matrix    map[string]*NodeInfo
	Workloads int
	Levels    int
}

type NodeInfo struct {
	Node
	WKL       string
	Exec      bool
	Reject    bool
	Candidate bool
	When      int
}

func GetMatrixKey(ID string, i int) (key string) {
	key = fmt.Sprintf("%v#%v", ID, i)
	return
}

func SplitMatrixKey(key string) (ID string, i int) {
	s := strings.Split(key, "#")
	if len(s) > 1 {
		wkl, _ := strconv.ParseInt(s[1], 0, 64)
		return s[0], int(wkl)
	} else {
		return "", -1
	}
}

func buildMatrix(wkls []string, nodes Nodes) (matrix NodesInfo) {
	iNodes := NodesInfo{make(map[string]*NodeInfo), -1, -1}
	max := -1
	for _, node := range nodes {
		for i, wkl := range wkls {
			iNodes.Matrix[GetMatrixKey(node.ID, i)] = &(NodeInfo{*node, wkl, false, false, false, -1})
		}
		if max < node.Level {
			max = node.Level
		}
	}
	iNodes.Workloads = len(wkls)
	iNodes.Levels = max
	return iNodes
}

func (c *Capacitor) HasMore(nodes *NodesInfo) bool {
	for _, node := range nodes.Matrix {
		if node.When == -1 {
			return true
		}
	}
	return false
}

func (c *Capacitor) NodesLeft(nodes *NodesInfo) (count int) {
	for _, node := range nodes.Matrix {
		if node.When == -1 {
			count++
		}
	}
	return
}

func (pNodes *NodesInfo) MarkCandidate(n *Node, metslo bool, exec int, cWKL int) {
	if n != nil {
		matrix := (*pNodes).Matrix
		for i := cWKL; i >= 0; i-- {
			key := GetMatrixKey(n.ID, i)
			nodeInfo := matrix[key]
			if nodeInfo.When == -1 {
				nodeInfo.Candidate = true
				nodeInfo.When = exec
				//fmt.Println("\tmarcando ", key, " candidato ")
			}
		}
		for _, node := range n.Lower {
			pNodes.MarkCandidate(node, metslo, exec, cWKL)
		}
	}
}

func (pNodes *NodesInfo) MarkReject(n *Node, metslo bool, exec int, cWKL int) {
	if n != nil {
		matrix := (*pNodes).Matrix
		for i := cWKL; i < pNodes.Workloads; i++ {
			key := GetMatrixKey(n.ID, i)
			nodeInfo := matrix[key]
			if nodeInfo.When == -1 {
				nodeInfo.Reject = true
				nodeInfo.When = exec
				//fmt.Println("\tmarcando ", key, " rejeitado ")
			}
		}
		for _, node := range n.Higher {
			pNodes.MarkReject(node, metslo, exec, cWKL)
		}
	}
}

func (pNodes *NodesInfo) Mark(key string, metslo bool, exec int) {
	_, cWKL := SplitMatrixKey(key)
	//fmt.Printf("INI MARK\n")
	//fmt.Printf("%v ? %v\n", key, metslo)
	matrix := (*pNodes).Matrix

	matrix[key].When = exec
	matrix[key].Exec = true

	if metslo {
		matrix[key].Candidate = true
		pNodes.MarkCandidate(&matrix[key].Node, metslo, exec, cWKL)
	} else {
		matrix[key].Reject = true
		pNodes.MarkReject(&matrix[key].Node, metslo, exec, cWKL)
	}
	//fmt.Printf("FIM MARK\n")
}

func (matrix NodesInfo) Clone() (clone *NodesInfo) {
	mapa := make(map[string]*NodeInfo)
	for key, node := range matrix.Matrix {
		mapa[key] = &(NodeInfo{node.Node, node.WKL, node.Exec, node.Reject, node.Candidate, node.When})
	}
	return &(NodesInfo{mapa, matrix.Workloads, matrix.Levels})
}

func (node NodeInfo) String() (str string) {
	return fmt.Sprintf("{when:%v, exec:%v, candidate:%v, reject:%v}", node.When, node.Exec, node.Candidate, node.Reject)
}

func (nodes NodesInfo) String() (str string) {
	str = fmt.Sprintf("{wkl:%v, heigh:%v, nodes:[\n\t", nodes.Workloads, nodes.Levels)
	for key, n := range nodes.Matrix {
		str = fmt.Sprintf("%v[%v]%v\n\t", str, key, n)
	}
	str = fmt.Sprintf("%v]}", str)
	return
}
