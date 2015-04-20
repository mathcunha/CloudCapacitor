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

func getMatrixKey(ID string, i int) (key string) {
	key = fmt.Sprintf("%v#%v", ID, i)
	return
}

func splitMatrixKey(key string) (ID string, i int) {
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
			n := new(NodeInfo)
			n.Node = *node
			n.WKL = wkl
			n.Exec = false
			n.Reject = false
			n.Candidate = false
			n.When = -1
			iNodes.matrix[getMatrixKey(node.ID, i)] = n
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

func (pNodes *NodesInfo) MarkCandidate(n *Node, metslo bool, exec int, cWKL int) {
	if n != nil {
		iNodes := *pNodes
		matrix := iNodes.matrix
		for i := cWKL; i >= 0; i-- {
			key := getMatrixKey(n.ID, i)
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
		iNodes := *pNodes
		matrix := iNodes.matrix
		for i := cWKL; i < pNodes.lenWKL; i++ {
			key := getMatrixKey(n.ID, i)
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
	_, cWKL := splitMatrixKey(key)
	//fmt.Printf("INI MARK\n")
	//fmt.Printf("%v ? %v\n", key, metslo)
	iNodes := *pNodes
	matrix := iNodes.matrix

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
