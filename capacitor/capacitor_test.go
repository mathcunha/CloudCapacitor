package capacitor

import (
	"log"
	"testing"
)

func TestMark(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}
	dspace := NewDeploymentSpace(&vms, 0.14, 2)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	c := Capacitor{dspace, m}
	mapa := c.Dspace.CapacityBy("Mem")
	M := *mapa

	matrix := buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["c3"])
	log.Println(M["c3"])
	log.Println(matrix)

	clone := matrix.Clone()
	clone.Mark("3.75_2", true, 1)
	clone.Mark("3.75_1", false, 2)
	log.Println(clone)

	matrix = buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["m3"])
	clone = matrix.Clone()
	log.Println(M["m3"])
	log.Println(matrix)
	clone.Mark("7.50_1", true, 1)
	clone.Mark("7.50_1", false, 2)
	log.Println(clone)
}
