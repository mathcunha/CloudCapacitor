package capacitor

import (
	"testing"
)

func TestLoadTypes(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	t.Logf("vms loaded %v", vms)
}
