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
	destDS := dspace.CapacityBy("Strict")
	_ = *dspace.CalcMaxSLO(e, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, []float32{10000, 20000, 30000, 40000, 50000})
	c, a, d := VerifyReflexionModel(dspace.Configurations(), destDS, false)

	t.Logf("DS Strict - convergence:%d, absence:%d, divergence:%d", c, a, d)

}

func TestReflexionModelDiffMem(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace_nocat.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	e := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	e.Load()
	destDS := dspace.CapacityBy("Mem")
	_ = *dspace.CalcMaxSLO(e, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, []float32{10000, 20000, 30000, 40000, 50000})
	c, a, d := VerifyReflexionModel(dspace.Configurations(), destDS, true)

	t.Logf("DS Mem - convergence:%d, absence:%d, divergence:%d", c, a, d)

}

func TestReflexionModelDiffPrice(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace_nocat.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	e := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	e.Load()
	destDS := dspace.CapacityBy("Price")
	_ = *dspace.CalcMaxSLO(e, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, []float32{10000, 20000, 30000, 40000, 50000})
	c, a, d := VerifyReflexionModel(dspace.Configurations(), destDS, true)

	t.Logf("DS Price - convergence:%d, absence:%d, divergence:%d", c, a, d)

}

func TestReflexionModelDiffCPU(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace_nocat.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	e := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	e.Load()
	destDS := dspace.CapacityBy("CPU")
	_ = *dspace.CalcMaxSLO(e, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, []float32{10000, 20000, 30000, 40000, 50000})
	c, a, d := VerifyReflexionModel(dspace.Configurations(), destDS, true)

	t.Logf("DS CPU - convergence:%d, absence:%d, divergence:%d", c, a, d)

}

func TestReflexionModelDiffCPUSafado(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace_nocat.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	e := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	e.Load()
	destDS := dspace.CapacityBy("CPU")

	c, a, d := DiffDS(dspace.Configurations(), destDS, destDS)
	t.Logf("DS CPU SAFADO- convergence:%d, absence:%d, divergence:%d", c, a, d)

}
