package main

import (
	"flag"
	"github.com/mathcunha/CloudCapacitor/capacitor"
	"log"
)

var (
	configFile string
	prop       string
	price      float64
	size       int
	routine    int
	slo        float64
)

func init() {
	flag.StringVar(&configFile, "config", "/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml", "vm types yml file")
	flag.Float64Var(&price, "price", 7.0, "configuration max price")
	flag.IntVar(&size, "size", 4, "max configuration size")
	flag.IntVar(&routine, "routine", 1000, "goroutine max number")
	flag.StringVar(&prop, "prop", "Mem", "[Mem,CPU,Price]")
	flag.Float64Var(&slo, "slo", 20000, "SLO")
}

func main() {
	flag.Parse()
	vms, err := capacitor.LoadTypes(configFile)
	if err != nil {
		log.Println("config error")
	}
	dspace := capacitor.NewDeploymentSpace(&vms, float32(price), size)
	m := capacitor.MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		log.Println("config error")
	}
	log.Printf("%v,%v,%v,%v\n", prop, slo, price, size)
	c := capacitor.Capacitor{dspace, m}
	h := capacitor.NewShortestPath(&c)
	h.Exec(prop, float32(slo), []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"})
}
