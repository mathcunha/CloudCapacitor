package capacitor

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

const (
	Conservative = "conservative"
	Pessimistic  = "pessimistic"
	Optimistic   = "optimistic"
	Hybrid       = "hybrid"
	Sensitive    = "sensitive"
	Adaptative   = "adaptative"
	Truth        = "truth"
)

const (
	Exec = iota
	Mark
)

type ExecInfo struct {
	Execs int
	Path  string
}

type NodeExec struct {
	nodes NodesInfo
	ExecInfo
}

type predictPoint struct {
	CapacitorPoint
	key       string
	nodesLeft int
	passed    bool
}
type byNodesLeft struct{ predictPoints []predictPoint }

func (s byNodesLeft) Len() int { return len(s.predictPoints) }
func (s byNodesLeft) Swap(i, j int) {
	s.predictPoints[i], s.predictPoints[j] = s.predictPoints[j], s.predictPoints[i]
}
func (s byNodesLeft) Less(i, j int) bool {
	return s.predictPoints[i].nodesLeft < s.predictPoints[j].nodesLeft
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
	c            *Capacitor
	equiBehavior int
}

//Explorer
type ExplorePath struct {
	c        *Capacitor
	maxExecs int
}

//the policies proposed at thesis
type Policy struct {
	c               *Capacitor
	levelPolicy     string
	wklPolicy       string
	equiBehavior    int
	isCapacityFirst bool
	useML           bool
}

func NewPolicy(c *Capacitor, levelPolicy string, wklPolicy string, equiBehavior int, isCapacityFirst bool, useML bool) (h *Policy) {
	switch levelPolicy {
	case Conservative, Pessimistic, Optimistic, Hybrid, Sensitive, Adaptative, Truth:
	default:
		log.Panicf("NewPolicy: levelPolicy not available:%v", levelPolicy)
	}
	switch wklPolicy {
	case Conservative, Pessimistic, Optimistic, Hybrid, Sensitive, Adaptative, Truth:
	default:
		log.Panicf("NewPolicy: wklPolicy not available:%v", wklPolicy)
	}
	switch equiBehavior {
	case Mark, Exec:
	default:
		log.Panicf("NewPolicy: equiBehavior not available:%v", equiBehavior)
	}
	return &(Policy{c, levelPolicy, wklPolicy, equiBehavior, isCapacityFirst, useML})
}

func NewShortestPath(c *Capacitor, equiBehavior int) (h *ShortestPath) {
	h = &ShortestPath{c, equiBehavior}
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

//how many nodes should left if key was executed
func whatIfNodesLeft(nodesInfo *NodesInfo, passed bool, key string, h *Policy, nodes Nodes, node *NodeInfo) int {
	localNodesInfo := nodesInfo.Clone()
	localNodesInfo.Mark(key, passed, 1, true)
	//Equivalents Actions
	if h.equiBehavior == Mark {
		equivalent := nodes.Equivalents(&node.Node)
		_, wkl := SplitMatrixKey(key)
		for _, lNode := range equivalent {
			lKey := GetMatrixKey(lNode.ID, wkl)
			lNodeInfo := nodesInfo.Matrix[lKey]
			if !(lNodeInfo.When != -1) {
				localNodesInfo.Mark(lKey, passed, 1, false)
			}
		}
	}
	return h.c.NodesLeft(localNodesInfo)
}

func (h *Policy) PredictNextNode(capPoints []CapacitorPoint, nodesInfo NodesInfo, nextKey string, slo float64, nodes Nodes) (key string, nodeInfo *NodeInfo) {
	//select by the Policy
	worstCase := whatIfNodesLeft(nodesInfo.Clone(), false, nextKey, h, nodes, nodesInfo.Matrix[nextKey])
	if nodesLeft := whatIfNodesLeft(nodesInfo.Clone(), true, nextKey, h, nodes, nodesInfo.Matrix[nextKey]); nodesLeft > worstCase {
		worstCase = nodesLeft
	}

	var predictPoints []predictPoint

	//calc mark potential
	for key, node := range nodesInfo.Matrix {
		if !(node.When != -1) {
			wkl, _ := strconv.Atoi(node.WKL)
			predictPoints = append(predictPoints, predictPoint{CapacitorPoint{config: *node.Config, wkl: wkl}, key, whatIfNodesLeft(nodesInfo.Clone(), false, key, h, nodes, node), false})
			predictPoints = append(predictPoints, predictPoint{CapacitorPoint{config: *node.Config, wkl: wkl}, key, whatIfNodesLeft(nodesInfo.Clone(), true, key, h, nodes, node), true})
		}
	}

	//asc order predictPoints
	sort.Sort(byNodesLeft{predictPoints})
	mapPrediction := make(map[string]float64)

	for _, v := range predictPoints {
		if v.nodesLeft >= worstCase {
			break
		}
		prediction, has := mapPrediction[v.key]
		if !has {
			prediction = Predict(capPoints, v.CapacitorPoint)
			mapPrediction[v.key] = prediction
		}
		if prediction > -1 {
			if v.passed {
				if prediction <= slo {
					key, nodeInfo = v.key, nodesInfo.Matrix[v.key]
					return
				}
			} else {
				if prediction > slo {
					key, nodeInfo = v.key, nodesInfo.Matrix[v.key]
					return
				}
			}
		}
	}

	return "", nil
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
					localNodesInfo.Mark(key, result.SLO <= slo, execInfo.Execs, true)
					//Equivalents Actions
					if h.equiBehavior == Mark {
						equivalent := nodes.Equivalents(&node.Node)
						_, wkl := SplitMatrixKey(key)
						for _, lNode := range equivalent {
							lKey := GetMatrixKey(lNode.ID, wkl)
							lNodeInfo := nodesInfo.Matrix[lKey]
							if !(lNodeInfo.When != -1) {
								localNodesInfo.Mark(lKey, result.SLO <= slo, execInfo.Execs, false)
							}
						}
					}
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
	nodesInfo.Mark(key, result.SLO <= slo, execInfo.Execs, true)
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
		var capPoints []CapacitorPoint
		nodes := (*dspace)[cat]
		dspaceInfo[cat] = buildMatrix(wkls, nodes)

		nodesInfo := dspaceInfo[cat]

		key := h.selectStartingPoint(&nodesInfo, &nodes)

		level := h.selectCapacityLevel(&nodesInfo, key, &nodes, nil, slo)
		wkl := h.selectWorkload(&nodesInfo, key, nil, slo)
		key, nodeInfo := h.NextConfig(&nodesInfo, nodes, level, wkl)

		//Process main loop, basically there will be no blank space
		for h.c.HasMore(&nodesInfo) {
			var result Result
			if nodeInfo.When == -1 {
				result = h.c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
				//log.Printf("[Policy.Exec] WKL:%v Result:%v\n", wkls[wkl], result)
				execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, key)
				execInfo.Execs++
				(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.Execs, true)
				if s, err := strconv.Atoi(nodeInfo.WKL); err == nil {
					capPoints = append(capPoints, CapacitorPoint{result.Config, s, float64(result.SLO)})
				}
			}

			//Equivalents Actions
			equivalent := nodes.Equivalents((&nodeInfo.Node))
			switch h.equiBehavior {
			case Exec:
				//execute all equivalents
				for _, node := range equivalent {
					key = GetMatrixKey(node.ID, wkl)
					nodeInfo = nodesInfo.Matrix[key]
					if !(nodeInfo.When != -1) {
						result = h.c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
						//log.Printf("[Policy.Exec] WKL:%v Result:%v\n", wkls[wkl], result)
						execInfo.Path = fmt.Sprintf("%v%v->", execInfo.Path, key)
						execInfo.Execs++
						(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.Execs, true)
						if s, err := strconv.Atoi(nodeInfo.WKL); err == nil {
							capPoints = append(capPoints, CapacitorPoint{result.Config, s, float64(result.SLO)})
						}
					}
				}
			case Mark:
				//mark all equivalents
				for _, node := range equivalent {
					key = GetMatrixKey(node.ID, wkl)
					nodeInfo = nodesInfo.Matrix[key]
					if !(nodeInfo.When != -1) {
						(&nodesInfo).Mark(key, result.SLO <= slo, execInfo.Execs, false)
					}
				}

			}

			if h.isCapacityFirst {
				//select capacity level
				oldLevel := level
				level = h.selectCapacityLevel(&nodesInfo, key, &nodes, &result, slo)

				//select workload
				if level == -1 {
					level = oldLevel
					wkl = h.selectWorkload(&nodesInfo, key, &result, slo)
				}
			} else {
				//select workload
				oldWKL := wkl
				wkl = h.selectWorkload(&nodesInfo, key, &result, slo)

				//select capacity level
				if wkl == -1 {
					wkl = oldWKL
					level = h.selectCapacityLevel(&nodesInfo, key, &nodes, &result, slo)
				}
			}

			//select other starting point
			if wkl == -1 || level == -1 {
				localKey := h.selectStartingPoint(&nodesInfo, &nodes)
				if localKey != "" {
					level = h.selectCapacityLevel(&nodesInfo, localKey, &nodes, &result, slo)
					wkl = h.selectWorkload(&nodesInfo, localKey, &result, slo)
				} else if h.c.HasMore(&nodesInfo) {
					log.Printf("ERROR: [Policy.Exec] Starting Point \n:%v\n%v", nodesInfo, nodes)
					break
				}
			}

			//next config
			if level != -1 {
				key, nodeInfo = h.NextConfig(&nodesInfo, nodes, level, wkl)
				if h.useML && h.c.HasMore(&nodesInfo) {
					if nodeInfo.When == -1 {
						if guessedKey, lNodeInfo := h.PredictNextNode(capPoints, nodesInfo, key, float64(slo), nodes); len(guessedKey) > 0 {
							fmt.Printf("picked %s, but my guess is %s\n", key, guessedKey)
							key, nodeInfo = guessedKey, lNodeInfo

						}
					} else {
						//find an equivalent
						equivalent := nodes.Equivalents((&nodeInfo.Node))
						for _, node := range equivalent {
							localKey := GetMatrixKey(node.ID, wkl)
							localNodeInfo := nodesInfo.Matrix[localKey]
							if !(localNodeInfo.When == -1) {
								if guessedKey, lNodeInfo := h.PredictNextNode(capPoints, nodesInfo, localKey, float64(slo), nodes); len(guessedKey) > 0 {
									fmt.Printf("picked %s, but my guess is %s\n", key, guessedKey)
									key, nodeInfo = guessedKey, lNodeInfo
									break
								}
							}
						}
					}
				}
			}
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
		if len(equivalent) > 0 {
			node := equivalent[0]
			if node.Config.Size < nodeInfo.Config.Size {
				key = GetMatrixKey(node.ID, wkl)
				nodeInfo = nodesInfo.Matrix[key]
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

func (p *Policy) selectWorkload(nodesInfo *NodesInfo, key string, result *Result, slo float32) (wklID int) {
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
	case Hybrid:
		wklPolicy := Pessimistic
		if result != nil {
			passed := result.SLO <= slo
			if !passed && result.CPU <= LowUsage && result.Mem <= LowUsage {
				//no configuration fits this slo
				wklPolicy = Pessimistic
			} else {
				if result.CPU >= HighUsage || result.Mem >= HighUsage {
					wklPolicy = Pessimistic
				} else if result.CPU <= LowUsage || result.Mem <= LowUsage {
					wklPolicy = Optimistic
				}
			}
		}
		policy := new(Policy)
		policy.wklPolicy = wklPolicy
		//log.Printf("hybrid WKL cpu:%v, mem:%v  choosing :%v", result.CPU, result.Mem, wklPolicy)
		return policy.selectWorkload(nodesInfo, key, result, slo)
	case Sensitive:
		policy := Optimistic
		if result != nil {
			passed := result.SLO <= slo
			delta := math.Abs(float64((result.SLO - slo) / slo))
			policy = Conservative
			if passed {
				if delta >= HighDelta {
					policy = Optimistic
				} else if delta <= LowDelta {
					policy = Pessimistic
				}
			} else {
				if delta >= HighDelta {
					policy = Pessimistic
				} else if delta <= LowDelta {
					policy = Optimistic
				}
			}

			//log.Printf("sensitive WKL key:%v passed:%v result:%.0f delta:%.4f choosing:%v", key, passed, result.SLO, delta, policy)
		}
		lPolicy := new(Policy)
		lPolicy.wklPolicy = policy
		return lPolicy.selectWorkload(nodesInfo, key, result, slo)
	case Adaptative:
		policy := Optimistic
		if result != nil {
			passed := result.SLO <= slo
			delta := math.Abs(float64((result.SLO - slo) / slo))
			policy = Conservative
			isHigh := result.CPU >= HighUsage || result.Mem >= HighUsage
			isLow := result.CPU <= LowUsage || result.Mem <= LowUsage
			if passed {
				if delta >= HighDelta {
					policy = Optimistic
					if isHigh {
						policy = Conservative
					}
				} else if delta <= LowDelta {
					policy = Pessimistic
					if isLow {
						policy = Conservative
					}
				}
			} else {
				if delta >= HighDelta {
					policy = Pessimistic
					if isLow {
						policy = Conservative
					}
				} else if delta <= LowDelta {
					policy = Optimistic
					if isHigh {
						policy = Conservative
					}
				}
			}
			//log.Printf("sensitive WKL key:%v passed:%v result:%.0f delta:%.4f choosing:%v", key, passed, result.SLO, delta, policy)
		}
		lPolicy := new(Policy)
		lPolicy.wklPolicy = policy
		return lPolicy.selectWorkload(nodesInfo, key, result, slo)
	case Truth:
		policy := Pessimistic
		if result != nil {
			passed := result.SLO <= slo
			delta := math.Abs(float64(result.SLO / slo))
			if !passed {
				if result.CPU <= LowUsage && result.Mem <= LowUsage {
					policy = Optimistic
				}
			}
			log.Printf("truth WKL key:%v passed:%v result:%.0f delta:%.4f cpu:%v mem:%v policy:%v", key, passed, result.SLO, delta, result.CPU, result.Mem, policy)
		}
		lPolicy := new(Policy)
		lPolicy.wklPolicy = policy
		return lPolicy.selectWorkload(nodesInfo, key, result, slo)
	}
	return
}

func (p *Policy) selectCapacityLevel(nodesInfo *NodesInfo, key string, nodes *Nodes, result *Result, slo float32) (level int) {
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
	case Hybrid:
		levelPolicy := Optimistic
		if result != nil {
			passed := result.SLO <= slo
			if !passed && result.CPU <= LowUsage && result.Mem <= LowUsage {
				//no configuration fits this slo
				levelPolicy = Pessimistic
			} else {
				if result.CPU >= HighUsage || result.Mem >= HighUsage {
					levelPolicy = Pessimistic
				} else if result.CPU <= LowUsage || result.Mem <= LowUsage {
					levelPolicy = Optimistic
				} else {
					levelPolicy = Conservative
				}
			}
		}
		policy := new(Policy)
		policy.levelPolicy = levelPolicy
		//log.Printf("hybrid LEVEL cpu:%v, mem:%v  choosing :%v", result.CPU, result.Mem, levelPolicy)
		return policy.selectCapacityLevel(nodesInfo, key, nodes, result, slo)
	case Sensitive:
		policy := Optimistic
		if result != nil {
			passed := result.SLO <= slo
			delta := math.Abs(float64((result.SLO - slo) / slo))
			policy = Conservative
			if passed {
				if delta >= HighDelta {
					policy = Optimistic
				} else if delta <= LowDelta {
					policy = Pessimistic
				}
			} else {
				if delta >= HighDelta {
					policy = Pessimistic
				} else if delta <= LowDelta {
					policy = Optimistic
				}
			}

			//log.Printf("sensitive LEVEL key:%v passed:%v result:%.0f delta:%.4f choosing:%v", key, passed, result.SLO, delta, policy)
		}
		lPolicy := new(Policy)
		lPolicy.levelPolicy = policy
		return lPolicy.selectCapacityLevel(nodesInfo, key, nodes, result, slo)
	case Adaptative:
		policy := Optimistic
		if result != nil {
			passed := result.SLO <= slo
			delta := math.Abs(float64((result.SLO - slo) / slo))
			policy = Conservative
			isHigh := result.CPU >= HighUsage || result.Mem >= HighUsage
			isLow := result.CPU <= LowUsage || result.Mem <= LowUsage
			if passed {
				if delta >= HighDelta {
					policy = Optimistic
					if isHigh {
						policy = Conservative
					}
				} else if delta <= LowDelta {
					policy = Pessimistic
					if isLow {
						policy = Conservative
					}
				}
			} else {
				if delta >= HighDelta {
					policy = Pessimistic
					if isLow {
						policy = Conservative
					}
				} else if delta <= LowDelta {
					policy = Optimistic
					if isHigh {
						policy = Conservative
					}
				}
			}
			//log.Printf("sensitive WKL key:%v passed:%v result:%.0f delta:%.4f choosing:%v", key, passed, result.SLO, delta, policy)
		}
		lPolicy := new(Policy)
		lPolicy.levelPolicy = policy
		return lPolicy.selectCapacityLevel(nodesInfo, key, nodes, result, slo)
	case Truth:
		policy := Optimistic
		if result != nil {
			passed := result.SLO <= slo
			delta := math.Abs(float64(result.SLO / slo))
			if !passed {
				if result.CPU <= LowUsage && result.Mem <= LowUsage {
					policy = Pessimistic
				}
			}
			log.Printf("truth level key:%v passed:%v result:%.0f delta:%.4f cpu:%v mem:%v policy:%v", key, passed, result.SLO, delta, result.CPU, result.Mem, policy)
		}
		lPolicy := new(Policy)
		lPolicy.levelPolicy = policy
		return lPolicy.selectCapacityLevel(nodesInfo, key, nodes, result, slo)

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
	//log.Printf(str)
}
