package capacitor

import (
	"testing"
)

func TestReflexionModelDiff(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	e := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	e.Load()
	srcDS := dspace.CapacityBy("Mem")
	dspace = *dspace.CalcMaxSLO(e, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, []float32{10000, 20000, 30000, 40000, 50000})
	destDS := dspace.CapacityBy("MaxSLO")
	DiffDS(dspace.Configurations(), srcDS, destDS)

}
