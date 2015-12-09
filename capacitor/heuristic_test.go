package capacitor

import (
	"testing"
)

func TestHeuristicConservative(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{dspace, m}
	h := NewPolicy(&c, Conservative, Conservative, 1, true)
	mode := "Strict"
	wkls := []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}
	execInfo, _ := h.Exec(mode, float32(20000), wkls)

	nodes := make([]*Node, 0, 28)

	for _, n := range *c.Dspace.CapacityBy(mode) {
		nodes = append(nodes, n...)
	}

	PrintExecPath(execInfo, wkls, nodes)
}

func TestHeuristicOptimistic(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{dspace, m}
	h := NewPolicy(&c, Optimistic, Optimistic, 1, true)
	mode := "Strict"
	wkls := []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}
	execInfo, _ := h.Exec(mode, float32(20000), wkls)

	nodes := make([]*Node, 0, 28)

	for _, n := range *c.Dspace.CapacityBy(mode) {
		nodes = append(nodes, n...)
	}

	PrintExecPath(execInfo, wkls, nodes)
}

func TestHeuristicPessimistic(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{dspace, m}
	h := NewPolicy(&c, Pessimistic, Pessimistic, 1, true)
	mode := "Strict"
	wkls := []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}
	execInfo, _ := h.Exec(mode, float32(20000), wkls)

	nodes := make([]*Node, 0, 28)

	for _, n := range *c.Dspace.CapacityBy(mode) {
		nodes = append(nodes, n...)
	}

	PrintExecPath(execInfo, wkls, nodes)
}
