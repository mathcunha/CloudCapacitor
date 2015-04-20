package capacitor

import (
	"testing"
)

func TestMark(t *testing.T) {
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
	if clone.matrix["3.75#0"].When != -1 {
		t.Fail()
	}
}
