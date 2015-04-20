package capacitor

import (
	"fmt"
	"github.com/mathcunha/CloudCapacitor/sync2"
	"log"
	"strings"
)

type ExecInfo struct {
	execs int
	path  string
	it    int
}

type NodeExec struct {
	nodes NodesInfo
	ExecInfo
}

type Heuristic interface {
	Exec(mode string, slo float32, wkls []string)
}

//Execute all configurations and workloads without infer
type BrutalForce struct {
	c *Capacitor
}

//Find the shortest path to Mark all configurations and workloads
type ShortestPath struct {
	c     *Capacitor
	slo   float32
	it    int
	maxIt int
}

func NewShortestPath(c *Capacitor) (h *ShortestPath) {
	h = new(ShortestPath)
	h.c = c
	return
}

func (bf *BrutalForce) Exec(mode string, slo float32, wkls []string) {
	mapa := bf.c.Dspace.CapacityBy(mode)
	for _, nodes := range *mapa {
		for _, node := range nodes {
			for _, conf := range node.Configs {
				for _, wkl := range wkls {
					result := bf.c.Executor.Execute(*conf, wkl)
					log.Printf("%v x %v ? %v \n", *conf, wkl, result.SLO <= slo)
				}
			}
		}
	}
}

func (h *ShortestPath) Exec(mode string, slo float32, wkls []string) {
	mapa := h.c.Dspace.CapacityBy(mode)
	h.slo = slo
	for _, nodes := range *mapa {
		h.ExecCategory(wkls, nodes)
	}
}

func (h *ShortestPath) ExecCategory(wkls []string, nodes Nodes) {
	numConfigs := 0
	for _, node := range nodes {
		numConfigs = numConfigs + len(node.Configs)
	}
	h.maxIt = len(wkls) * numConfigs

	nexts := []NodeExec{NodeExec{buildMatrix(wkls, nodes), ExecInfo{0, "", 0}}}

	for i := 0; i <= h.maxIt; i++ {
		wg := sync2.NewBlockWaitGroup(100000)
		chBest := make(chan ExecInfo)

		wg.Add(1)
		go func() {
			defer wg.Done()
			nexts = h.findShortestPath(nexts, wg, chBest, h.maxIt)
		}()

		go func() {
			wg.Wait()
			close(chBest)
		}()

		best := h.GetBest(chBest)

		if best.execs != -1 {
			PrintExecPath(best, wkls, nodes)
			return
		}

	}

}

func PrintExecPath(winner ExecInfo, wkls []string, nodes Nodes) {
	path := strings.Split(winner.path, "->")
	str := ""
	execs := 0
	for _, key := range path {
		ID, cWKL := splitMatrixKey(key)
		if cWKL != -1 {
			node := nodes.NodeByID(ID)
			str = fmt.Sprintf("%vWorkload:%v, Configs:%v\n", str, wkls[cWKL], node.Configs)
			execs = execs + len(node.Configs)
		}
	}
	str = fmt.Sprintf("%vTotal Execs:%v", str, execs)
	log.Printf(str)
}

func (h *ShortestPath) findShortestPath(current []NodeExec, wg *sync2.BlockWaitGroup, chBest chan ExecInfo, numConfigs int) (nexts []NodeExec) {
	novo := new([]NodeExec)
	bingo := false
	nexts = *novo
	lessNodes := numConfigs
	for _, ex := range current {
		for key, node := range ex.nodes.matrix {
			if !(node.When != -1) {
				cNodes := ex.nodes.Clone()
				nExecs := ex.execs
				result := Result{}
				for _, conf := range node.Configs {
					nExecs = nExecs + 1

					result = h.c.Executor.Execute(*conf, node.WKL)

					cNodes.Mark(key, result.SLO <= h.slo, nExecs)

				}
				nPath := fmt.Sprintf("%v%v->", ex.path, key)
				//c.Exec(*cNodes, slo, nExecs, nPath, wg, ch, it+1, maxIts)

				if h.c.HasMore(cNodes) {
					if !bingo {
						leftNodes := h.c.NodesLeft(cNodes)
						if lessNodes == leftNodes {
							nEx := new(NodeExec)
							nEx.nodes = *cNodes
							nEx.execs = nExecs
							nEx.path = nPath
							nEx.it = ex.it + 1
							nexts = append(nexts, *nEx)
						}
						if lessNodes > leftNodes {
							lessNodes = leftNodes
							nEx := new(NodeExec)
							nEx.nodes = *cNodes
							nEx.execs = nExecs
							nEx.path = nPath
							nEx.it = ex.it + 1
							nexts = []NodeExec{*nEx}
						}
					}
				} else {
					//All executions!
					bingo = true
					wg.Add(1)
					go func() {
						defer wg.Done()
						chBest <- ExecInfo{nExecs, nPath, -1}
					}()
					//return nil
				}
			}
		}
	}

	return nexts
}

func (h *ShortestPath) GetBest(chBest chan ExecInfo) (best ExecInfo) {
	best = ExecInfo{-1, "", -1}
	for {
		execInfo, more := <-chBest
		if more {
			best = execInfo
		} else {
			break
		}
	}

	return best
}
