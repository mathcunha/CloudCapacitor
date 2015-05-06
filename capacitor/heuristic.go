package capacitor

import (
	"fmt"
	"github.com/mathcunha/CloudCapacitor/sync2"
	"log"
	"strings"
)

const (
	Conservative = "conservative"
	Pessimist    = "pessimist"
	Optimist     = "optimist"
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

//the policies proposed at thesis
type Policy struct {
	c           *Capacitor
	levelPolicy string
	wklPolicy   string
}

func NewPolicy(c *Capacitor, levelPolicy string, wklPolicy string) (h *Policy) {
	return &(Policy{c, levelPolicy, wklPolicy})
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

func (h *Policy) Exec(mode string, slo float32, wkls []string) {
	dspace := h.c.Dspace.CapacityBy(mode)

	execs := 0

	//map to store the results by category
	dspaceInfo := make(map[string]NodesInfo)

	for cat, nodes := range *dspace {
		log.Printf("[Policy.Exec] Category:%v\n", cat)
		dspaceInfo[cat] = buildMatrix(wkls, nodes)

		wkl := h.selectWorkload(false, dspace, "", cat)
		level := h.selectCapacityLevel(false, dspace, "", cat)
		nodesInfo := dspaceInfo[cat]

		key := getMatrixKey(nodes.NodeByLevel(level).ID, wkl)
		nodeInfo := nodesInfo.matrix[key]

		//Process main loop, basically there will be no blank space
		for h.c.HasMore(&nodesInfo) {
			result := h.c.Executor.Execute(*nodeInfo.Configs[0], nodeInfo.WKL)
			log.Printf("[Policy.Exec] Node: %v - Result :%v\n", nodeInfo, result)
			execs++
			metSLO := result.SLO <= slo
			(&nodesInfo).Mark(key, metSLO, execs)

			log.Printf("[Policy.Exec] loop :%v\n", nodesInfo)
			//execute all equivalents
			equivalent := nodeInfo.Node.Equivalents()
			for _, node := range equivalent {
				key = getMatrixKey(node.ID, wkl)
				nodeInfo = nodesInfo.matrix[key]
				if !(nodeInfo.When != -1) {
					result = h.c.Executor.Execute(*nodeInfo.Configs[0], nodeInfo.WKL)
					log.Printf("[Policy.Exec] Node: %v - Result :%v\n", nodeInfo, result)
					execs++
					metSLO = result.SLO <= slo
					(&nodesInfo).Mark(key, metSLO, execs)
				}
			}
			//execute all equivalents
			level = h.selectCapacityLevel(metSLO, dspace, key, cat)

			//select workload
			if level == -1 {
				wkl = h.selectCapacityLevel(metSLO, dspace, key, cat)
			}

			//select other starting point

			break
		}
	}
}

func HasMore(c *Capacitor, dspaceInfo map[string]NodesInfo) (hasMore bool) {
	hasMore = true
	for _, nodes := range dspaceInfo {
		hasMore = c.HasMore(&nodes) || hasMore
	}
	return
}

func (p *Policy) selectWorkload(metSLO bool, mapa *map[string]Nodes, key string, cat string) (wklID int) {
	if "" == key {
		return 0
	}
	//TODO
	return -1
}

func (p *Policy) selectCapacityLevel(metSLO bool, mapa *map[string]Nodes, key string, cat string) (level int) {
	if "" == key {
		return 1
	}
	//TODO
	return -1
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
	nexts = *(new([]NodeExec))
	lessNodes := numConfigs
	for _, ex := range current {
		for key, node := range ex.nodes.matrix {
			if !(node.When != -1) {
				cNodes := ex.nodes.Clone()
				nExecs := ex.execs
				var result Result
				for _, conf := range node.Configs {
					nExecs = nExecs + 1
					result = h.c.Executor.Execute(*conf, node.WKL)
					cNodes.Mark(key, result.SLO <= h.slo, nExecs)

				}
				nPath := fmt.Sprintf("%v%v->", ex.path, key)

				if nodesLeft := h.c.NodesLeft(cNodes); nodesLeft != 0 {
					if lessNodes == nodesLeft {
						nexts = append(nexts, NodeExec{*cNodes, ExecInfo{nExecs, nPath, ex.it + 1}})
					}
					if lessNodes > nodesLeft {
						lessNodes = nodesLeft
						nexts = []NodeExec{NodeExec{*cNodes, ExecInfo{nExecs, nPath, ex.it + 1}}}
					}
				} else {
					//All executions!
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
