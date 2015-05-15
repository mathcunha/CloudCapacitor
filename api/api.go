package api

import (
	"encoding/json"
	"fmt"
	"github.com/mathcunha/CloudCapacitor/capacitor"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

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

func callCapacitorResource(w http.ResponseWriter, r *http.Request) {
	var config struct {
		Slo           int     `json:"slo"`
		Price         float32 `json:"price"`
		Size          int     `json:"instances"`
		Mode          string  `json:"mode"`
		Category      bool    `json:"category"`
		Demand        []int   `json:"demand"`
		WKL           string  `json:"wkl"`
		Configuration string  `json:"configuration"`
		Heuristic     string  `json:"heuristic"`
	}

	err := json.NewDecoder(r.Body).Decode(&config)
	r.Body.Close()
	if err != nil {
		log.Printf("ERROR: Parsing request body callCapacitorResource:%v", err)
		return
	}

	vms, err := capacitor.LoadTypes("config/dspace.yml")
	if err != nil {
		log.Println("ERROR: callCapacitorResource loanding vm types")
		return
	}
	dspace := capacitor.NewDeploymentSpace(&vms, config.Price, config.Size)
	m := capacitor.MockExecutor{"config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		log.Println("ERROR: callCapacitorResource unable to start MockExecutor")
		return
	}

	c := capacitor.Capacitor{dspace, m}
	var h capacitor.Heuristic
	switch config.Heuristic {
	case "bf":
		h = capacitor.NewBrutalForce(&c)
	case "sp":
		h = capacitor.NewShortestPath(&c)
	default:
		h = capacitor.NewPolicy(&c, config.Configuration, config.WKL)
	}

	wkls := make([]string, len(config.Demand), len(config.Demand))
	for i, d := range config.Demand {
		wkls[i] = strconv.Itoa(d)
	}

	execInfo, dspaceInfo := h.Exec(config.Mode, float32(config.Slo), wkls)

	str := ExecPathSummary(execInfo, wkls, config.Mode, &c)
	str = fmt.Sprintf("%v, \"spaceInfo\":%v}", str[0:len(str)-1], DeploymentSpace(execInfo, wkls, config.Mode, &c, dspaceInfo, m, float32(config.Slo)))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%v", str)
}

func DeploymentSpace(winner capacitor.ExecInfo, wkls []string, mode string, c *capacitor.Capacitor, dspaceInfo map[string]capacitor.NodesInfo, executor capacitor.Executor, slo float32) (str string) {
	nodeMap := *c.Dspace.CapacityBy(mode)

	str = "["
	for _, cat := range c.Dspace.Categories() {
		nodesInfo := dspaceInfo[cat]
		nodes := nodeMap[cat]
		for level := 1; level <= nodesInfo.Levels; level++ {
			node := nodes.NodeByLevel(level)
			for wkl := 0; wkl < nodesInfo.Workloads; wkl++ {
				str = fmt.Sprintf("%v%v", str, PrintNodeExecInfo(node, wkl, nodesInfo, executor, slo))
			}
			for _, e := range nodes.Equivalents(node) {
				for wkl := 0; wkl < nodesInfo.Workloads; wkl++ {
					str = fmt.Sprintf("%v%v", str, PrintNodeExecInfo(e, wkl, nodesInfo, executor, slo))
				}
			}
		}
	}
	str = fmt.Sprintf("%v]", str[0:len(str)-1])
	return
}

func PrintNodeExecInfo(node *capacitor.Node, wkl int, nodesInfo capacitor.NodesInfo, executor capacitor.Executor, slo float32) (str string) {
	key := capacitor.GetMatrixKey(node.ID, wkl)
	nodeInfo := nodesInfo.Matrix[key]
	if wkl == 0 {
		str = fmt.Sprintf("%v{\"name\":\"%v\", \"size\":%v, \"wkl\":[", str, node.Config.Name, node.Config.Size)
	}
	result := executor.Execute(*node.Config, nodeInfo.WKL)
	metSLO := result.SLO <= slo
	right := nodeInfo.Candidate == metSLO
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
	str = fmt.Sprintf("{\"execs\":%v, \"price\":%.2f, \"path\":%v]}", execs, price, str[0:len(str)-1])
	return str
}
