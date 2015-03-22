package capacitor

import (
	"fmt"
	"github.com/mathcunha/CloudCapacitor/sync2"
	"log"
)

type ExecInfo struct {
	Execs int
	Path  string
}

type currentExec struct {
	nodes NodesInfo
	key   string
	execs int
	path  string
	it    int
}

type nextExec struct {
	nodes NodesInfo
	execs int
	path  string
	it    int
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
	wg    *sync2.BlockWaitGroup
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
	for key, nodes := range *mapa {
		h.ExecCategory(wkls, nodes)
		log.Println("Category[", key, "] - ", nodes)
	}
}

func (h *ShortestPath) InitControlers() {
	h.wg = sync2.NewBlockWaitGroup(100000)
}

func (h *ShortestPath) PostExecs(nexts []nextExec) (current []currentExec) {
	for _, next := range nexts {
		for key, _ := range next.nodes.matrix {
			current = append(current, currentExec{next.nodes, key, next.execs, next.path, next.it})
		}
	}
	return
}

func (h *ShortestPath) ExecCategory(wkls []string, nodes Nodes) {
	numConfigs := 0
	for _, node := range nodes {
		numConfigs = numConfigs + len(node.Configs)
	}
	h.maxIt = len(wkls) * numConfigs
	log.Printf("Max iterations :%v \n", h.maxIt)

	h.InitControlers()

	nexts := []nextExec{nextExec{buildMatrix(wkls, nodes), 0, "", 1}}
	var best ExecInfo

	for i := 1; i <= h.maxIt; i++ {
		log.Printf("Now trying %v Iteration(s) \n", i)

		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			nexts, best = h.findShortestPath(h.PostExecs(nexts))
		}()

		h.wg.Wait()

		if nexts == nil {
			log.Printf("the winner is %v", best)
			return
		}

	}

}

func (h *ShortestPath) findShortestPath(current []currentExec) (nexts []nextExec, shortest ExecInfo) {
	for _, ex := range current {
		node := ex.nodes.matrix[ex.key]
		if !(node.When != -1) {
			cNodes := ex.nodes.Clone()
			nExecs := ex.execs
			result := Result{}
			for _, conf := range node.Configs {
				nExecs = nExecs + 1

				result = h.c.Executor.Execute(*conf, node.WKL)

				cNodes.Mark(ex.key, result.SLO <= h.slo, nExecs)

			}
			nPath := fmt.Sprintf("%v%v->", ex.path, ex.key)
			//c.Exec(*cNodes, slo, nExecs, nPath, wg, ch, it+1, maxIts)

			if h.c.HasMore(cNodes) {
				nEx := new(nextExec)
				nEx.nodes = *cNodes
				nEx.execs = nExecs
				nEx.path = nPath
				nEx.it = ex.it + 1
				nexts = append(nexts, *nEx)
			} else {
				//All executions!
				return nil, ExecInfo{nExecs, nPath}
			}
		}
	}

	return nexts, ExecInfo{-1, ""}
}

func (h *ShortestPath) GetBest() (best ExecInfo) {
	best = ExecInfo{-1, ""}

	/*	for len(h.chEndExec) > 0 {
		execInfo := <-h.chEndExec
		best = execInfo
		log.Printf("%v, %v \n", execInfo.Execs, execInfo.Path)
	}*/

	return best
}
