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
	clone.Mark("3.75#2", true, 1)
	clone.Mark("3.75#1", false, 2)
	if c.NodesLeft(clone) != 0 {
		t.Fail()
	}

	matrix = buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["m3"])
	clone = matrix.Clone()
	clone.Mark("7.50#1", true, 1)
	clone.Mark("7.50#1", false, 2)
	if clone.Matrix["3.75#0"].When != -1 {
		t.Fail()
	}
}
