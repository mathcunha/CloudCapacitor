package capacitor

import (
	"testing"
)

func TestNewDeploymentSpace(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	t.Logf("vms loaded %v", vms)
	dspace := NewDeploymentSpace(&vms, 7.0, 4)

	t.Logf("configs generated %v", dspace)
}

func TestCapacityCPU(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	dspace.CapacityBy("cpu")
	t.Logf("configs by cpu", dspace)
}

func TestCapacityMem(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	dspace.CapacityBy("mem")
	t.Logf("configs by mem", dspace)
}

func TestCapacityPrice(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	dspace.CapacityBy("price")
	t.Logf("configs by price", dspace)
}
