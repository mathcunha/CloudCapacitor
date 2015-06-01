package capacitor

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
)

const (
	Conservative = "conservative"
	Pessimistic  = "pessimistic"
	Optimistic   = "optimistic"
)

type ExecInfo struct {
	Execs int
	Path  string
}

type NodeExec struct {
	nodes NodesInfo
	ExecInfo
}

type Heuristic interface {
	Exec(mode string, slo float32, wkls []string) (ExecInfo, map[string]NodesInfo)
}

//Execute all configurations and workloads without infer
type BrutalForce struct {
	c *Capacitor
}

//Find the shortest path to Mark all configurations and workloads
type ShortestPath struct {
	c *Capacitor
}

//Explorer
type ExplorePath struct {
	c        *Capacitor
	maxExecs int
}

//the policies proposed at thesis
type Policy struct {
	c           *Capacitor
	levelPolicy string
	wklPolicy   string
}

func NewPolicy(c *Capacitor, levelPolicy string, wklPolicy string) (h *Policy) {
	switch levelPolicy {
	case Conservative, Pessimistic, Optimistic:
	default:
		log.Panicf("NewPolicy: levelPolicy not available:%v", levelPolicy)
	}
	switch wklPolicy {
	case Conservative, Pessimistic, Optimistic:
	default:
		log.Panicf("NewPolicy: wklPolicy not available:%v", wklPolicy)
	}
	return &(Policy{c, levelPolicy, wklPolicy})
}

func NewShortestPath(c *Capacitor) (h *ShortestPath) {
	h = &ShortestPath{c}
	return
}

func NewBrutalForce(c *Capacitor) (h *BrutalForce) {
	h = &BrutalForce{c}
	return
}

func NewExplorePath(c *Capacitor, maxExecs int) (h *ExplorePath) {
	h = &ExplorePath{c, maxExecs}
	return
}

func (h *BrutalForce) Exec(mode string, slo float32, wkls []string) (path ExecInfo, dspaceInfo map[string]NodesInfo) {
	dspace := h.c.Dspace.CapacityBy(mode)
	execInfo := ExecInfo{0, ""}

	//map to store the results by category
	dspaceInfo = make(map[string]NodesInfo)
	for _, cat := range h.c.Dspace.Categories() {
		nodes := (*dspace)[cat]
		nodesInfo := buildMatrix(wkls, nodes)
		for key, node := range nodesInfo.Matrix {
			result := h.c.Executor.Execute(*node.Config, node.WKL)
			execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, key)
			execInfo.Execs++
			node.When = execInfo.Execs
			node.Exec = true
			if result.SLO <= slo {
				node.Candidate = true
			} else {
				node.Reject = true
			}
		}
		dspaceInfo[cat] = nodesInfo
	}
	return execInfo, dspaceInfo
}

func (h *ShortestPath) Exec(mode string, slo float32, wkls []string) (path ExecInfo, dspaceInfo map[string]NodesInfo) {
	dspace := h.c.Dspace.CapacityBy(mode)
	execInfo := ExecInfo{0, ""}

	//map to store the results by category
	dspaceInfo = make(map[string]NodesInfo)

	for _, cat := range h.c.Dspace.Categories() {
		nodes := (*dspace)[cat]
		nodesInfo := buildMatrix(wkls, nodes)

		for h.c.HasMore(&nodesInfo) {
			execInfo.Execs++
			bestKey := ""
			var bestNodesInfo NodesInfo
			nodesLeft := math.MaxInt32
			for key, node := range nodesInfo.Matrix {
				if !(node.When != -1) {
					localNodesInfo := nodesInfo.Clone()
					result := h.c.Executor.Execute(*node.Config, node.WKL)
					localNodesInfo.Mark(key, result.SLO <= slo, execInfo.Execs)
					if localNodesLeft := h.c.NodesLeft(localNodesInfo); localNodesLeft != 0 {
						if nodesLeft > localNodesLeft {
							nodesLeft = localNodesLeft
							bestNodesInfo = *localNodesInfo
							bestKey = key
						}
					} else {
						dspaceInfo[cat] = *localNodesInfo
						bestKey = key
						bestNodesInfo = *localNodesInfo
						break
					}
				}
			}
			nodesInfo = bestNodesInfo
			execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, bestKey)
		}
	}

	return execInfo, dspaceInfo
}

func (h *ExplorePath) Exec(mode string, slo float32, wkls []string) (path ExecInfo, dspaceInfo map[string]NodesInfo) {
	dspace := h.c.Dspace.CapacityBy(mode)
	//map to store the results by category
	dspaceInfo = make(map[string]NodesInfo)
	execInfo := ExecInfo{0, ""}

	for _, cat := range h.c.Dspace.Categories() {
		nodes := (*dspace)[cat]
		nodesInfo := buildMatrix(wkls, nodes)
		for i := 1; i <= h.maxExecs; i++ {
			for key, _ := range nodesInfo.Matrix {
				if localExecInfo, localNodesInfo := h.Explore(slo, key, nodesInfo.Clone(), ExecInfo{execInfo.Execs, execInfo.Path}, i); localExecInfo.Execs != -1 {
					dspaceInfo[cat] = *localNodesInfo
					execInfo = localExecInfo
					i = h.maxExecs + 1
					break
				}
			}
		}

	}

	return execInfo, dspaceInfo
}

func (h *ExplorePath) Explore(slo float32, key string, nodesInfo *NodesInfo, execInfo ExecInfo, maxExecs int) (ExecInfo, *NodesInfo) {
	execInfo.Execs++
	execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, key)
	nodeInfo := nodesInfo.Matrix[key]
	result := h.c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
	nodesInfo.Mark(key, result.SLO <= slo, execInfo.Execs)
	if h.c.HasMore(nodesInfo) {
		if execInfo.Execs < maxExecs {
			for localKey, _ := range nodesInfo.Matrix {
				if localExecInfo, localNodesInfo := h.Explore(slo, localKey, nodesInfo.Clone(), ExecInfo{execInfo.Execs, execInfo.Path}, maxExecs); localExecInfo.Execs != -1 {
					return localExecInfo, localNodesInfo
				}
			}
		}
	} else {
		return execInfo, nodesInfo
	}

	//not finished
	return ExecInfo{-1, ""}, nodesInfo
}

func (h *Policy) Exec(mode string, slo float32, wkls []string) (path ExecInfo, dspaceInfo map[string]NodesInfo) {
	dspace := h.c.Dspace.CapacityBy(mode)

	execInfo := ExecInfo{0, ""}

	//map to store the results by category
	dspaceInfo = make(map[string]NodesInfo)

	for _, cat := range h.c.Dspace.Categories() {
		nodes := (*dspace)[cat]
		dspaceInfo[cat] = buildMatrix(wkls, nodes)

		nodesInfo := dspaceInfo[cat]

		key := h.selectStartingPoint(&nodesInfo, &nodes)

		level := h.selectCapacityLevel(&nodesInfo, key, &nodes)
		wkl := h.selectWorkload(&nodesInfo, key)
		key = GetMatrixKey(nodes.NodeByLevel(level).ID, wkl)
		nodeInfo := nodesInfo.Matrix[key]

		//Process main loop, basically there will be no blank space
		for h.c.HasMore(&nodesInfo) {
			if !(nodeInfo.When != -1) {
				result := h.c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
				//log.Printf("[Policy.Exec] WKL:%v Result:%v\n", wkls[wkl], result)
				execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, key)
				execInfo.Execs++
				(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.Execs)
			}

			//execute all equivalents
			equivalent := nodes.Equivalents((&nodeInfo.Node))
			for _, node := range equivalent {
				key = GetMatrixKey(node.ID, wkl)
				nodeInfo = nodesInfo.Matrix[key]
				if !(nodeInfo.When != -1) {
					result := h.c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
					//log.Printf("[Policy.Exec] WKL:%v Result:%v\n", wkls[wkl], result)
					execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, key)
					execInfo.Execs++
					(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.Execs)
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
				localKey := h.selectStartingPoint(&nodesInfo, &nodes)
				if localKey != "" {
					level = h.selectCapacityLevel(&nodesInfo, localKey, &nodes)
					wkl = h.selectWorkload(&nodesInfo, localKey)
				} else if h.c.HasMore(&nodesInfo) {
					log.Fatalf("[Policy.Exec] Starting Point \n:%v\n%v", nodesInfo, nodes)
					break
				}
			}

			//next config
			key, nodeInfo = h.NextConfig(&nodesInfo, nodes, level, wkl)
		}
		//log.Printf("[Policy.Exec] Category:%v Execs:%v", cat, execInfo.execs)
	}

	return execInfo, dspaceInfo
}

func (p *Policy) NextConfig(nodesInfo *NodesInfo, nodes Nodes, level int, wkl int) (key string, nodeInfo *NodeInfo) {
	key = GetMatrixKey(nodes.NodeByLevel(level).ID, wkl)
	nodeInfo = nodesInfo.Matrix[key]

	if nodeInfo != nil {
		//it is ordered
		equivalent := nodes.Equivalents((&nodeInfo.Node))
		for _, node := range equivalent {
			localKey := GetMatrixKey(node.ID, wkl)
			localNodeInfo := nodesInfo.Matrix[localKey]
			if localNodeInfo.Config.Size < nodeInfo.Config.Size {
				nodeInfo = localNodeInfo
				key = localKey
			}
		}
	}
	return
}

func HasMore(c *Capacitor, dspaceInfo map[string]NodesInfo) (hasMore bool) {
	hasMore = true
	for _, nodes := range dspaceInfo {
		hasMore = c.HasMore(&nodes) || hasMore
	}
	return
}

func (p *Policy) selectStartingPoint(nodesInfo *NodesInfo, nodes *Nodes) (key string) {
	for level := 1; level <= nodesInfo.Levels; level++ {
		for wkl := 0; wkl < nodesInfo.Workloads; wkl++ {
			key = GetMatrixKey(nodes.NodeByLevel(level).ID, wkl)
			nodeInfo := nodesInfo.Matrix[key]
			if nodeInfo.When == -1 {
				return
			} else {
				//Same level, but, possibly, different lowers and highers
				equivalents := nodes.Equivalents((&nodeInfo.Node))
				for _, node := range equivalents {
					key = GetMatrixKey(node.ID, wkl)
					nodeInfo := nodesInfo.Matrix[key]
					if nodeInfo.When == -1 {
						return
					}
				}
			}
		}
	}
	return ""
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
	wkls = make([]int, 0, nodesInfo.Workloads)
	ID, wkl := SplitMatrixKey(key)
	for i := 0; i < nodesInfo.Workloads; i++ {
		nodeInfo := nodesInfo.Matrix[GetMatrixKey(ID, i)]
		if nodeInfo.When == -1 {
			wkls = append(wkls, i)
		}
	}
	sort.Ints(wkls)
	return
}

//capacity levels availables in the current workload
func (p *Policy) buildCapacityLevelList(key string, nodesInfo *NodesInfo, nodes *Nodes) (ID string, levels []int) {
	levels = make([]int, 0, nodesInfo.Levels)
	ID, wkl := SplitMatrixKey(key)
	for i := 1; i <= nodesInfo.Levels; i++ {
		nodeInfo := nodesInfo.Matrix[GetMatrixKey(nodes.NodeByLevel(i).ID, wkl)]
		if nodeInfo.When == -1 {
			levels = append(levels, i)
		} else {
			//Same level, but, possibly, different lowers and highers
			equivalents := nodes.Equivalents((&nodeInfo.Node))
			for _, node := range equivalents {
				nodeInfo := nodesInfo.Matrix[GetMatrixKey(node.ID, wkl)]
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

func PrintExecPath(winner ExecInfo, wkls []string, nodes Nodes) {
	path := strings.Split(winner.Path, "->")
	str := ""
	execs := 0
	for _, key := range path {
		ID, cWKL := SplitMatrixKey(key)
		if cWKL != -1 {
			node := nodes.NodeByID(ID)
			str = fmt.Sprintf("%v{Workload:%v, Level:%v, Config:%v}\n", str, wkls[cWKL], node.Level, node.Config)
			execs++
		}
	}
	str = fmt.Sprintf("%vTotal Execs:%v", str, execs)
	log.Printf(str)
}
