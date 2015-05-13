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
		http.Error(w, "method not allowed, try curl -X POST https://cloudcapacitor.herokuapp.com/api/v1/capacitor", http.StatusMethodNotAllowed)
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
	price := 7.0
	size := 4
	var config struct {
		Slo           int    `json:"slo"`
		Mode          string `json:"mode"`
		Category      bool   `json:"category"`
		Demand        []int  `json:"demand"`
		WKL           string `json:"wkl"`
		Configuration string `json:"configuration"`
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
	dspace := capacitor.NewDeploymentSpace(&vms, float32(price), size)
	m := capacitor.MockExecutor{"config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		log.Println("ERROR: callCapacitorResource unable to start MockExecutor")
		return
	}

	c := capacitor.Capacitor{dspace, m}
	h := capacitor.NewPolicy(&c, config.Configuration, config.WKL)

	wkls := make([]string, len(config.Demand), len(config.Demand))
	for i, d := range config.Demand {
		wkls[i] = strconv.Itoa(d)
	}

	execInfo := h.Exec(config.Mode, float32(config.Slo), wkls)

	str := ExecPathSummary(execInfo, wkls, config.Mode, &c)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%v", str)
}

func ExecPathSummary(winner capacitor.ExecInfo, wkls []string, mode string, c *capacitor.Capacitor) (str string) {
	nos := make([]*capacitor.Node, 0, 28)

	for _, n := range *c.Dspace.CapacityBy(mode) {
		nos = append(nos, n...)
	}
	nodes := capacitor.Nodes(nos)
	execs := 0
	price := float32(0.0)

	path := strings.Split(winner.Path, "->")
	str = "["
	for _, key := range path {
		ID, cWKL := capacitor.SplitMatrixKey(key)
		if cWKL != -1 {
			node := nodes.NodeByID(ID)
			str = fmt.Sprintf("%v{\"key\":\"%v\", \"workload\":%v, \"level\":%v,  \"name\":\"%v\", \"price\":%.2f, \"size\":%v},", str, key, wkls[cWKL], node.Level, node.Configs[0].Name, node.Configs[0].Price(), node.Configs[0].Size)
			execs = execs + len(node.Configs)
			price = price + node.Configs[0].Price()
		}
	}
	//one extra comma
	str = fmt.Sprintf("{\"execs\":%v, \"price\":%.2f, \"path\":%v]}", execs, price, str[0:len(str)-1])
	return str
}
