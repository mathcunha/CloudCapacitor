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
