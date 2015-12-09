package api

import (
	"encoding/json"
	"fmt"
	"github.com/mathcunha/CloudCapacitor/capacitor"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type ExecInfo struct {
	Config *capacitor.Configuration
	right  bool
}

func restHandlerV1(w http.ResponseWriter, r *http.Request) {
	a_path := strings.Split(r.URL.Path, "/")
	if "POST" != r.Method {
		http.Error(w, "method not allowed, try curl -d '{\"slo\":20000, \"demand\":[100], \"wkl\":\"optimistic\", \"configuration\":\"optimistic\", \"mode\":\"Strict\"}' -X POST https://cloudcapacitor.herokuapp.com/api/v1/capacitor", http.StatusMethodNotAllowed)
		return
	}
	if "capacitor" != a_path[3] {
		http.Error(w, "no handler to path "+r.URL.Path, http.StatusNotFound)
		return
	} else {
		callCapacitorResource(w, r)
	}
}

func StartServer() {
	http.HandleFunc("/webui/", staticHandler)
	http.HandleFunc("/api/v1/", restHandlerV1)
	http.ListenAndServe(getPort(), nil)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	//log.Printf("INFO: staticHandler:%v", r.URL.Path[1:])
	http.ServeFile(w, r, r.URL.Path[1:])
}

func getPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "8080"
		log.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

func NodesToDOT(r *http.Request) (graph string) {
	var config struct {
		Slo      float32 `json:"slo"`
		Price    float32 `json:"price"`
		Size     int     `json:"instances"`
		Mode     string  `json:"mode"`
		Category bool    `json:"category"`
		VMtype   []int   `json:"vmtype"`
		Demand   []int   `json:"demand"`
		App      string  `json:"app"`
	}
	err := json.NewDecoder(r.Body).Decode(&config)
	r.Body.Close()
	if err != nil {
		log.Printf("ERROR: Parsing request body drawDeploymentSpace:%v", err)
		return
	}
	file := "config/dspace.yml"
	if !config.Category {
		file = "config/dspace_nocat.yml"
	}

	vms, err := getVMTypes(file, config.VMtype)
	if err != nil {
		log.Println("ERROR: drawDeploymentSpace loanding vm types")
		return
	}
	dspace := capacitor.NewDeploymentSpace(&vms, config.Price, config.Size)
	if config.Mode == "MaxSLO" {
		m := capacitor.MockExecutor{"config/" + config.App + "_cpu_mem.csv", nil}
		err = m.Load()
		if err != nil {
			log.Println("ERROR: callCapacitorResource unable to start MockExecutor")
			return
		}

		wkls := make([]string, len(config.Demand), len(config.Demand))
		for i, d := range config.Demand {
			wkls[i] = strconv.Itoa(d)
		}
		slos := []float32{10000, 20000, 30000, 40000, 50000}
		if config.App == "terasort" {
			slos = []float32{0.01, 0.02, 0.03, 0.04, 0.05}
		}
		dspace = *dspace.CalcMaxSLO(m, wkls, slos)
	}
	graph = capacitor.NodesToDOT(dspace.CapacityBy(config.Mode))
	return
}

func drawDeploymentSpace(w http.ResponseWriter, r *http.Request) {
	graph := NodesToDOT(r)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%v", callGraphviz(graph))
}

func deploymentSpaceAsDOT(w http.ResponseWriter, r *http.Request) {
	graph := NodesToDOT(r)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%v", graph)
}

func callGraphviz(graph string) string {
	resp, err := http.PostForm("http://graphviz-dev.appspot.com/create_preview", url.Values{"engine": {"dot"}, "script": {graph}})
	if err != nil {
		log.Println("ERROR: drawDeploymentSpace calling remote service: %v", err)
		return "ERROR"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR: drawDeploymentSpace reading remote service: %v", err)
		return "ERROR"
	}
	return string(body)
}

func getVMTypes(file string, excludedVMs []int) (vms []capacitor.VM, err error) {
	vms, err = capacitor.LoadTypes(file)
	if err != nil {
		return
	}
	if len(excludedVMs) > 0 {
		localVMs := make([]capacitor.VM, len(vms)-len(excludedVMs), len(vms)-len(excludedVMs))
		j := 0
		k := 0
		for i := 0; i < len(vms); i++ {
			if j == len(excludedVMs) || i != excludedVMs[j] {
				localVMs[k] = vms[i]
				k++
			} else {
				j++
			}
		}
		vms = localVMs
	}
	for i, _ := range vms {
		vms[i].Strict = i
	}
	return
}

func callCapacitorResource(w http.ResponseWriter, r *http.Request) {
	a_path := strings.Split(r.URL.Path, "/")
	if "dot" == a_path[4] {
		deploymentSpaceAsDOT(w, r)
	} else if "draw" == a_path[4] {
		drawDeploymentSpace(w, r)
	} else {
		var config struct {
			Slo             float32 `json:"slo"`
			Price           float32 `json:"price"`
			Size            int     `json:"instances"`
			Mode            string  `json:"mode"`
			Category        bool    `json:"category"`
			Demand          []int   `json:"demand"`
			VMtype          []int   `json:"vmtype"`
			WKL             string  `json:"wkl"`
			Configuration   string  `json:"configuration"`
			Heuristic       string  `json:"heuristic"`
			MaxExecs        int     `json:"maxExecs"`
			EquiBehavior    int     `json:"equiBehavior"`
			App             string  `json:"app"`
			IsCapacityFirst bool    `json:"isCapacityFirst"`
		}

		err := json.NewDecoder(r.Body).Decode(&config)
		r.Body.Close()
		if err != nil {
			log.Printf("ERROR: Parsing request body callCapacitorResource:%v", err)
			return
		}

		file := "config/dspace.yml"
		if !config.Category {
			file = "config/dspace_nocat.yml"
		}

		vms, err := getVMTypes(file, config.VMtype)
		if err != nil {
			log.Println("ERROR: callCapacitorResource loanding vm types")
			return
		}
		dspace := capacitor.NewDeploymentSpace(&vms, config.Price, config.Size)
		m := capacitor.MockExecutor{"config/" + config.App + "_cpu_mem.csv", nil}
		err = m.Load()
		if err != nil {
			log.Println("ERROR: callCapacitorResource unable to start MockExecutor")
			return
		}

		wkls := make([]string, len(config.Demand), len(config.Demand))
		for i, d := range config.Demand {
			wkls[i] = strconv.Itoa(d)
		}

		if config.Mode == "MaxSLO" {
			slos := []float32{10000, 20000, 30000, 40000, 50000}
			if config.App == "terasort" {
				slos = []float32{0.01, 0.02, 0.03, 0.04, 0.05}
			}
			dspace = *dspace.CalcMaxSLO(m, wkls, slos)
		}

		c := capacitor.Capacitor{dspace, m}
		var h capacitor.Heuristic
		switch config.Heuristic {
		case "e":
			h = capacitor.NewExplorePath(&c, config.MaxExecs)
		case "bf":
			h = capacitor.NewBrutalForce(&c)
		case "sp":
			h = capacitor.NewShortestPath(&c, config.EquiBehavior)
		default:
			h = capacitor.NewPolicy(&c, config.Configuration, config.WKL, config.EquiBehavior, config.IsCapacityFirst)
		}

		execInfo, dspaceInfo := h.Exec(config.Mode, config.Slo, wkls)
		bestPrice := make(map[string]ExecInfo)

		str := ExecPathSummary(execInfo, wkls, config.Mode, &c)
		if strDeployment := DeploymentSpace(config.Mode, &c, dspaceInfo, m, config.Slo, &bestPrice); len(strDeployment) > 1 {
			str = fmt.Sprintf("%v, \"spaceInfo\":%v}", str[0:len(str)-1], strDeployment)
		} else {
			str = fmt.Sprintf("%v, \"spaceInfo\":[]}", str[0:len(str)-1])
		}

		success := 0
		for _, e := range bestPrice {
			if e.right {
				success += 1
			}
		}

		str = fmt.Sprintf("%v,%v, \"success\":%v}", str[0:len(str)-1], ExecsByKey(execInfo, &c, dspaceInfo, wkls), float32(success)/float32(len(config.Demand)))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%v", str)
	}
}

func CalcAccurary(dspaceInfo map[string]capacitor.NodesInfo, executor capacitor.Executor, slo float32) (int, int, int) {
	tp, fp, fn := 0, 0, 0

	for _, nodesInfo := range dspaceInfo {
		for _, nodeInfo := range nodesInfo.Matrix {
			result := executor.Execute(*nodeInfo.Node.Config, nodeInfo.WKL)
			metSLO := result.SLO <= slo

			if nodeInfo.Candidate && metSLO {
				tp += 1
			}

			if nodeInfo.Candidate && !metSLO {
				fp += 1
			}

			if nodeInfo.Reject && metSLO {
				fn += 1
			}
		}
	}

	return tp, fp, fn

}

//http://en.wikipedia.org/wiki/Precision_and_recall
func CalcFmeasure(tp int, fp int, fn int) (fmeasure float64) {
	precision := float64(tp) / float64(tp+fp)
	recall := float64(tp) / float64(tp+fn)

	fmeasure = (2 * (precision * recall) / (precision + recall))

	if math.IsNaN(fmeasure) {
		log.Printf("tp:%v,fp:%v,fn:%v,precision:%v,recall:%v", tp, fp, fn, precision, recall)
		return 0
	}

	return
}

func DeploymentSpace(mode string, c *capacitor.Capacitor, dspaceInfo map[string]capacitor.NodesInfo, executor capacitor.Executor, slo float32, bestPrice *map[string]ExecInfo) (str string) {
	nodeMap := *c.Dspace.CapacityBy(mode)

	str = "["
	for _, cat := range c.Dspace.Categories() {
		nodesInfo := dspaceInfo[cat]
		nodes := nodeMap[cat]
		for level := 1; level <= nodesInfo.Levels; level++ {
			node := nodes.NodeByLevel(level)
			for wkl := 0; wkl < nodesInfo.Workloads; wkl++ {
				str = fmt.Sprintf("%v%v", str, PrintNodeExecInfo(node, wkl, nodesInfo, executor, slo, bestPrice))
			}
			for _, e := range nodes.Equivalents(node) {
				for wkl := 0; wkl < nodesInfo.Workloads; wkl++ {
					str = fmt.Sprintf("%v%v", str, PrintNodeExecInfo(e, wkl, nodesInfo, executor, slo, bestPrice))
				}
			}
		}
	}
	str = fmt.Sprintf("%v], \"fmeasure\":%.4f ", str[0:len(str)-1], CalcFmeasure(CalcAccurary(dspaceInfo, executor, slo)))

	//log.Printf(str)
	return
}

func PrintNodeExecInfo(node *capacitor.Node, wkl int, nodesInfo capacitor.NodesInfo, executor capacitor.Executor, slo float32, bestPrice *map[string]ExecInfo) (str string) {
	key := capacitor.GetMatrixKey(node.ID, wkl)
	nodeInfo := nodesInfo.Matrix[key]
	if wkl == 0 {
		str = fmt.Sprintf("%v{\"name\":\"%v\", \"size\":%v, \"wkl\":[", str, node.Config.Name, node.Config.Size)
	}
	result := executor.Execute(*node.Config, nodeInfo.WKL)
	metSLO := result.SLO <= slo
	right := nodeInfo.Candidate == metSLO
	if metSLO {
		e, has := (*bestPrice)[nodeInfo.WKL]
		if has {
			if e.Config == nil || e.Config.Price() > (*node).Config.Price() {
				(*bestPrice)[nodeInfo.WKL] = ExecInfo{node.Config, right}
			}
		} else {
			(*bestPrice)[nodeInfo.WKL] = ExecInfo{node.Config, right}
		}
	} else {
		_, has := (*bestPrice)[nodeInfo.WKL]
		if !has {
			(*bestPrice)[nodeInfo.WKL] = ExecInfo{nil, right}
		}
	}
	str = fmt.Sprintf("%v{\"wkl\":\"%v\", \"when\":%v, \"exec\":%v, \"cadidate\":%v, \"right\":%v},", str, nodeInfo.WKL, nodeInfo.When, nodeInfo.Exec, nodeInfo.Candidate, right)
	if wkl == nodesInfo.Workloads-1 {
		str = fmt.Sprintf("%v]},", str[0:len(str)-1])
	}
	return
}

func ExecPathSummary(winner capacitor.ExecInfo, wkls []string, mode string, c *capacitor.Capacitor) (str string) {
	nos := make([]*capacitor.Node, 0, 28)

	for _, n := range *c.Dspace.CapacityBy(mode) {
		nos = append(nos, n...)
	}
	nodes := capacitor.Nodes(nos)
	execs := 0
	price := 0.0

	path := strings.Split(winner.Path, "->")
	str = "["
	for _, key := range path {
		ID, cWKL := capacitor.SplitMatrixKey(key)
		if cWKL != -1 {
			node := nodes.NodeByID(ID)
			str = fmt.Sprintf("%v{\"key\":\"%v\", \"workload\":%v, \"level\":%v,  \"name\":\"%v\", \"price\":%.2f, \"size\":%v},", str, key, wkls[cWKL], node.Level, node.Config.Name, node.Config.Price(), node.Config.Size)
			execs++
			price = price + float64(node.Config.Price())
		}
	}
	//one extra comma
	if len(str) > 1 {
		str = fmt.Sprintf("{\"execs\":%v, \"price\":%.2f, \"path\":%v]}", execs, price, str[0:len(str)-1])
	} else {
		str = fmt.Sprintf("{\"execs\":%v, \"price\":%.2f, \"path\":[]}", execs, price)
	}
	return str
}

func ExecsByKey(winner capacitor.ExecInfo, c *capacitor.Capacitor, dspaceInfo map[string]capacitor.NodesInfo, wkls []string) (str string) {
	path := strings.Split(winner.Path, "->")
	execsByKey := make([]int, len(path), len(path))

	//initializing
	for i, _ := range execsByKey {
		execsByKey[i] = -1
	}

	for _, cat := range c.Dspace.Categories() {
		nodesInfo := dspaceInfo[cat]
		for _, nodeInfo := range nodesInfo.Matrix {
			execsByKey[nodeInfo.When]++
		}
	}

	str = "["

	for i, key := range path {
		ID, iWKL := capacitor.SplitMatrixKey(key)
		if key != "" {
			str = fmt.Sprintf("%v{\"key\":\"%v\", \"execs\":%v},", str, ID+"("+wkls[iWKL]+")", execsByKey[i+1])
		}
	}

	if len(str) > 1 {
		str = fmt.Sprintf("\"execsByKey\":%v]", str[0:len(str)-1])
	} else {
		str = "\"execsByKey\":[]"
	}

	return
}
