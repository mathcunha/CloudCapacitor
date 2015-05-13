package api

import (
	"encoding/json"
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
		slo           int
		mode          string
		category      bool
		demand        []int
		wkl           string
		configuration string
	}

	err := json.NewDecoder(r.Body).Decode(&config)
	r.Body.Close()
	if err != nil {
		log.Printf("callCapacitorResource:%v", err)
		return
	}

	vms, err := capacitor.LoadTypes("../config/dspace.yml")
	if err != nil {
		log.Println("callCapacitorResource config error")
	}
	dspace := capacitor.NewDeploymentSpace(&vms, float32(price), size)
	m := capacitor.MockExecutor{"../config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		log.Println("config error")
	}

	log.Printf("%v,%v,%v\n", config, price, size)
	c := capacitor.Capacitor{dspace, m}
	h := capacitor.NewPolicy(&c, config.wkl, config.configuration)

	wkl := make([]string, len(config.demand), len(config.demand))
	for i, d := range config.demand {
		wkl[i] = strconv.Itoa(d)
	}

	h.Exec(config.mode, float32(config.slo), wkl)
}
