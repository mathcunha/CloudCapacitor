package capacitor

import (
	"fmt"
	"github.com/mathcunha/CloudCapacitor/sync2"
	"log"
	"sort"
	"strings"
)

const (
	Conservative = "conservative"
	Pessimistic  = "pessimistic"
	Optimistic   = "optimistic"
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
	Exec(mode string, slo float32, wkls []string) ExecInfo
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

func (bf *BrutalForce) Exec(mode string, slo float32, wkls []string) (path ExecInfo) {
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
	//TODO
	return ExecInfo{0, "", 0}
}

func (h *ShortestPath) Exec(mode string, slo float32, wkls []string) (path ExecInfo) {
	mapa := h.c.Dspace.CapacityBy(mode)
	h.slo = slo
	for _, nodes := range *mapa {
		h.ExecCategory(wkls, nodes)
	}
	//TODO
	return ExecInfo{0, "", 0}
}

func (h *Policy) Exec(mode string, slo float32, wkls []string) (path ExecInfo) {
	dspace := h.c.Dspace.CapacityBy(mode)

	execInfo := ExecInfo{0, "", 0}

	//map to store the results by category
	dspaceInfo := make(map[string]NodesInfo)

	for cat, nodes := range *dspace {
		dspaceInfo[cat] = buildMatrix(wkls, nodes)

		nodesInfo := dspaceInfo[cat]

		level, wkl := h.selectStartingPoint(&nodesInfo, &nodes)
		key := getMatrixKey(nodes.NodeByLevel(level).ID, wkl)

		level = h.selectCapacityLevel(&nodesInfo, key, &nodes)
		wkl = h.selectWorkload(&nodesInfo, key)
		key = getMatrixKey(nodes.NodeByLevel(level).ID, wkl)
		nodeInfo := nodesInfo.matrix[key]

		//Process main loop, basically there will be no blank space
		for h.c.HasMore(&nodesInfo) {
			if !(nodeInfo.When != -1) {
				result := h.c.Executor.Execute(*nodeInfo.Configs[0], nodeInfo.WKL)
				//log.Printf("[Policy.Exec] WKL:%v Result:%v\n", wkls[wkl], result)
				execInfo.path = fmt.Sprintf("%v%v->", execInfo.path, key)
				execInfo.execs++
				(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.execs)
			}

			//execute all equivalents
			equivalent := nodes.Equivalents((&nodeInfo.Node))
			for _, node := range equivalent {
				key = getMatrixKey(node.ID, wkl)
				nodeInfo = nodesInfo.matrix[key]
				if nodeInfo == nil {
					log.Printf("ERROR: [Policy.Exec] %v", key)
					return execInfo
				}
				if !(nodeInfo.When != -1) {
					result := h.c.Executor.Execute(*nodeInfo.Configs[0], nodeInfo.WKL)
					//log.Printf("[Policy.Exec] WKL:%v Result:%v\n", wkls[wkl], result)
					execInfo.path = fmt.Sprintf("%v%v->", execInfo.path, key)
					execInfo.execs++
					(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.execs)
				}
			}
			//select capacity level
			oldLevel := level
			level = h.selectCapacityLevel(&nodesInfo, key, &nodes)

			//select workload
			if level == -1 {
				level = oldLevel
				wkl = h.selectWorkload(&nodesInfo, key)
			}

			//select other starting point
			if wkl == -1 {
				level, wkl = h.selectStartingPoint(&nodesInfo, &nodes)
				if wkl != -1 && level != -1 {
					localKey := getMatrixKey(nodes.NodeByLevel(level).ID, wkl)
					level = h.selectCapacityLevel(&nodesInfo, localKey, &nodes)
					wkl = h.selectWorkload(&nodesInfo, localKey)
				} else if h.c.HasMore(&nodesInfo) {
					log.Fatalf("[Policy.Exec] Starting Point \n:%v\n%v", nodesInfo, nodes)
				}
			}

			//next config
			if wkl != -1 && level != -1 {
				key = getMatrixKey(nodes.NodeByLevel(level).ID, wkl)
				nodeInfo = nodesInfo.matrix[key]
			}
		}
		//log.Printf("[Policy.Exec] Category:%v Execs:%v", cat, execInfo.execs)
	}

	return execInfo
}

func HasMore(c *Capacitor, dspaceInfo map[string]NodesInfo) (hasMore bool) {
	hasMore = true
	for _, nodes := range dspaceInfo {
		hasMore = c.HasMore(&nodes) || hasMore
	}
	return
}

func (p *Policy) selectStartingPoint(nodesInfo *NodesInfo, nodes *Nodes) (level int, wkl int) {
	for level = 1; level <= nodesInfo.levels; level++ {
		for wkl = 0; wkl < nodesInfo.workloads; wkl++ {
			nodeInfo := nodesInfo.matrix[getMatrixKey(nodes.NodeByLevel(level).ID, wkl)]
			if nodeInfo.When == -1 {
				return
			} else {
				//Same height, but, possibly, different lowers and highers
				equivalents := nodes.Equivalents((&nodeInfo.Node))
				for _, node := range equivalents {
					nodeInfo := nodesInfo.matrix[getMatrixKey(node.ID, wkl)]
					if nodeInfo.When == -1 {
						return
					}
				}
			}
		}
	}
	return -1, -1
}

func (p *Policy) selectWorkload(nodesInfo *NodesInfo, key string) (wklID int) {
	_, wkls := p.buildWorkloadList(key, nodesInfo)
	wklID = -1
	if len(wkls) == 0 {
		return
	}

	switch p.wklPolicy {
	case Conservative:
		wklID = wkls[len(wkls)/2]
	case Pessimistic:
		wklID = wkls[0]
	case Optimistic:
		wklID = wkls[len(wkls)-1]
	}
	return
}

func (p *Policy) selectCapacityLevel(nodesInfo *NodesInfo, key string, nodes *Nodes) (level int) {
	_, levels := p.buildCapacityLevelList(key, nodesInfo, nodes)
	level = -1
	if len(levels) == 0 {
		return
	}

	switch p.levelPolicy {
	case Conservative:
		level = levels[len(levels)/2]
	case Optimistic:
		level = levels[0]
	case Pessimistic:
		level = levels[len(levels)-1]

	}
	return
}

//Workloads availables in the current capacity level
func (p *Policy) buildWorkloadList(key string, nodesInfo *NodesInfo) (wkl int, wkls []int) {
	wkls = make([]int, 0, nodesInfo.workloads)
	ID, wkl := splitMatrixKey(key)
	for i := 0; i < nodesInfo.workloads; i++ {
		nodeInfo := nodesInfo.matrix[getMatrixKey(ID, i)]
		if nodeInfo.When == -1 {
			wkls = append(wkls, i)
		}
	}
	sort.Ints(wkls)
	return
}

//capacity levels availables in the current workload
func (p *Policy) buildCapacityLevelList(key string, nodesInfo *NodesInfo, nodes *Nodes) (ID string, levels []int) {
	levels = make([]int, 0, nodesInfo.levels)
	ID, wkl := splitMatrixKey(key)
	for i := 1; i <= nodesInfo.levels; i++ {
		nodeInfo := nodesInfo.matrix[getMatrixKey(nodes.NodeByLevel(i).ID, wkl)]
		if nodeInfo.When == -1 {
			levels = append(levels, i)
		} else {
			//Same height, but, possibly, different lowers and highers
			equivalents := nodes.Equivalents((&nodeInfo.Node))
			for _, node := range equivalents {
				nodeInfo := nodesInfo.matrix[getMatrixKey(node.ID, wkl)]
				if nodeInfo.When == -1 {
					levels = append(levels, i)
					break
				}
			}
		}
	}
	sort.Ints(levels)
	return
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
			str = fmt.Sprintf("%v{Workload:%v, Level:%v, Configs:%v}\n", str, wkls[cWKL], node.Level, node.Configs)
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
