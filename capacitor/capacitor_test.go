package capacitor

import (
	"testing"
)

func TestBF(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	dspace := NewDeploymentSpace(&vms, 0.15, 1)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{dspace, m}
	c.MinExec("Mem", 70000, []string{"100", "200", "300", "400"})
}
