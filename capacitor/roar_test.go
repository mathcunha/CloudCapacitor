package capacitor

import (
	"testing"
)

func TestRoar(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	e := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/terasort_cpu_mem.csv", nil}
	e.Load()
	throughtput, _ := NewROARExecutor(&vms, []string{"10000000", "20000000", "30000000", "40000000", "50000000"}, e)
	t.Logf("ROAR generated %v", throughtput)
}
