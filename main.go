package main

import (
	"github.com/mathcunha/CloudCapacitor/capacitor"
	"log"
)

func main() {
	vms, err := capacitor.LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		log.Println("config error")
	}
	dspace := capacitor.NewDeploymentSpace(&vms, 0.14, 1)
	m := capacitor.MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		log.Println("config error")
	}
	c := capacitor.Capacitor{dspace, m}
	c.MinExec("Mem", 70000, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"})
}
