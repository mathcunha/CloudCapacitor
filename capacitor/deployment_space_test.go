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
	dspace.CapacityBy("CPU")
	//t.Logf("configs by cpu", dspace)
	//t.Logf("mapa by cpu", mapa)
}

func TestCapacityMem(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	mapa := dspace.CapacityBy("Mem")
	t.Logf("mapa by mem :%v", printTree(mapa))
}

/*func TestCapacityPrice(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	mapa := dspace.CapacityBy("Price")
	t.Logf("configs by price", dspace)
	t.Logf("mapa by price", printTree(mapa))
}*/
