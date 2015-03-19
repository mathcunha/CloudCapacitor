package capacitor

import (
	"testing"
)

func TestBF(t *testing.T) {
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
	c.MinExec("Mem", 20000, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"})
}
