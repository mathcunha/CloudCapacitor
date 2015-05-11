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

func TestCapacityStrict(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	mapa := dspace.CapacityBy("Strict")
	t.Logf("configs generated %v", dspace)
	t.Logf("mapa by Strict :%v", printTree(mapa))
	nodes := (*mapa)["c3"]

	for _, n := range nodes {
		if n.ID == "3_c3_large" {
			t.Logf("\n %v equivalents", n.ID)
			for _, e := range nodes.Equivalents(n) {
				t.Logf("%v,", e.ID)
			}
			t.Log("\n")
		}
		if n.ID == "2_c3_xlarge" {
			t.Logf("\n %v equivalents", n.ID)
			for _, e := range nodes.Equivalents(n) {
				t.Logf("%v,", e.ID)
			}
			t.Log("\n")
		}
		if n.ID == "2_c3_2xlarge" {
			t.Logf("\n %v equivalents", n.ID)
			for _, e := range nodes.Equivalents(n) {
				t.Logf("%v,", e.ID)
			}
			t.Log("\n")
		}

	}

	nodes = (*mapa)["m3"]
	for _, n := range nodes {
		if n.ID == "4_m3_medium" {
			t.Logf("\n %v equivalents", n.ID)
			for _, e := range nodes.Equivalents(n) {
				t.Logf("%v,", e.ID)
			}
			t.Log("\n")
		}
	}
}
