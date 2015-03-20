package capacitor

import (
	"log"
	"sync"
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

	matrix := buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["m3"])
	log.Println(M["m3"])
	log.Println(matrix)

	clone := matrix.Clone()
	clone.Mark("2_4", false, 1)

	if !clone.matrix["2_4"].Exec {
		t.Log("2_4 not marked as Exec")
		t.Fail()
	}

	if !clone.matrix["2_5"].Reject {
		t.Log("2_5 not marked as Reject")
		t.Fail()
	}

	clone.Mark("1_3", true, 2)

	if !clone.matrix["1_3"].Exec {
		t.Log("1_3 not marked as Exec")
		t.Fail()
	}

	if !clone.matrix["1_2"].Candidate {
		t.Log("1_2 not marked as Candidate")
		t.Fail()
	}

	clone.Mark("1_3", true, 5)

	//c.MinExec("Mem", 70000, []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"})
}

func TestExec(t *testing.T) {
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

	matrix := buildMatrix([]string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000"}, M["m3"])
	log.Println(M["m3"])
	log.Println(matrix)

	clone := matrix.Clone()
	clone.Mark("2_4", false, 1)
	clone.Mark("1_3", true, 2)

	clone.matrix["1_5"].When = -1
	clone.matrix["1_5"].Candidate = false
	clone.matrix["2_4"].When = -1
	clone.matrix["2_4"].Reject = false
	clone.matrix["2_5"].When = -1
	clone.matrix["2_5"].Reject = false

	wg := new(sync.WaitGroup)
	wg.Add(1)
	ch := make(chan ExecInfo)
	go func() {
		defer wg.Done()

		c.Exec(*clone, 170000, 0, "", wg, ch)
	}()
	go func() {
		wg.Wait()
		close(ch)
	}()
	best := c.WaitExec(ch)

	log.Printf("the winner is :%v", best)
}
