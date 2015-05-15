package capacitor

import (
	"testing"
)

func TestMarkStrict(t *testing.T) {
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
	mapa := c.Dspace.CapacityBy("Strict")
	M := *mapa

	matrix := buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["c3"])

	clone := matrix.Clone()
	clone.Mark("1_c3_large#0", true, 1)

	if !clone.Matrix["1_c3_2xlarge#0"].Candidate {
		t.Fail()
	}
}

func TestMarkMemShortestPath(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	d := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{d, m}
	mode := "Mem"
	dspace := c.Dspace.CapacityBy(mode)
	cat := "c3"

	nodesInfo := buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, (*dspace)[cat])
	//keys := []string{"3_c3_2xlarge#4", "3_c3_large#3", "2_c3_large#2", "4_c3_2xlarge#5", "1_c3_large#1", "1_c3_large#0", "4_c3_2xlarge#4"}
	keys := []string{"3_c3_2xlarge#4", "3_c3_large#3"}
	execs := 0
	for _, key := range keys {
		execs++
		nodeInfo := nodesInfo.Matrix[key]
		result := c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
		(&nodesInfo).Mark(key, result.SLO <= 10000, execs)
	}

	t.Logf("%v - BEFORE KEY[2_c3_large#2]", c.NodesLeft(&nodesInfo))
	nodeInfo := nodesInfo.Matrix["2_c3_large#2"]
	result := c.Executor.Execute(*nodeInfo.Config, nodeInfo.WKL)
	(&nodesInfo).Mark("2_c3_large#2", result.SLO <= 10000, execs)
	t.Logf("%v - AFTER KEY[2_c3_large#2]", c.NodesLeft(&nodesInfo))
}

func TestMarkMem(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	dspace := NewDeploymentSpace(&vms, 0.14, 2)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{dspace, m}
	mapa := c.Dspace.CapacityBy("Mem")
	M := *mapa

	matrix := buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["c3"])

	clone := matrix.Clone()
	clone.Mark("1_c3_large#2", true, 1)
	clone.Mark("1_c3_large#1", false, 2)
	if c.NodesLeft(clone) != 0 {
		t.Fail()
	}

	matrix = buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["m3"])
	clone = matrix.Clone()
	clone.Mark("2_m3_medium#1", true, 1)
	clone.Mark("2_m3_medium#1", false, 2)
	if clone.Matrix["1_m3_medium#0"].When != -1 {
		t.Fail()
	}
}
