package capacitor

import (
	"testing"
)

func TestCapacityMem(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	mapa := dspace.CapacityBy("Mem")
	t.Logf("configs generated %v", dspace)
	t.Logf("mapa by mem :%v", printTree(mapa))
}
