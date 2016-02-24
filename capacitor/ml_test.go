package capacitor

import (
	"testing"
)

func TestMachineLearningWorkload(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/terasort_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	//wkls := []string{"10000000", "20000000", "30000000", "40000000", "50000000"}
	configs := (*dspace.configs)["c3"]
	i := 0
	capPoints := make([]CapacitorPoint, 3, 3)

	result := m.Execute(*configs[0], "10000000")
	capPoints[i] = CapacitorPoint{result.Config, 10000000, float64(result.SLO)}
	i++

	result = m.Execute(*configs[len(configs)-1], "10000000")
	capPoints[i] = CapacitorPoint{result.Config, 10000000, float64(result.SLO)}
	i++

	result = m.Execute(*configs[len(configs)/2], "10000000")
	capPoints[i] = CapacitorPoint{result.Config, 10000000, float64(result.SLO)}
	i++

	Predict(capPoints, CapacitorPoint{config: *configs[1], wkl: 10000000}, 100000)
}

func TestMachineLearningConfiguration(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/terasort_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	//wkls := []string{"10000000", "20000000", "30000000", "40000000", "50000000"}
	configs := (*dspace.configs)["c3"]
	i := 0
	capPoints := make([]CapacitorPoint, 3, 3)

	result := m.Execute(*configs[0], "10000000")
	capPoints[i] = CapacitorPoint{result.Config, 10000000, float64(result.SLO)}
	i++

	result = m.Execute(*configs[0], "20000000")
	capPoints[i] = CapacitorPoint{result.Config, 20000000, float64(result.SLO)}
	i++

	result = m.Execute(*configs[0], "30000000")
	capPoints[i] = CapacitorPoint{result.Config, 30000000, float64(result.SLO)}
	i++

	Predict(capPoints, CapacitorPoint{config: *configs[0], wkl: 40000000}, 100000)
}

func TestMachineLearningConfigurationMiddleXLarge(t *testing.T) {
	vms, err := LoadTypes("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/dspace.yml")
	if err != nil {
		t.Errorf("config error")
	}

	dspace := NewDeploymentSpace(&vms, 7.0, 4)
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/terasort_cpu_mem.csv", nil}
	err = m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	//wkls := []string{"10000000", "20000000", "30000000", "40000000", "50000000"}
	configs := (*dspace.configs)["c3"]
	i := 0
	capPoints := make([]CapacitorPoint, 2, 2)

	result := m.Execute(*configs[2], "10000000")
	capPoints[i] = CapacitorPoint{result.Config, 10000000, float64(result.SLO)}
	i++

	result = m.Execute(*configs[2], "30000000")
	capPoints[i] = CapacitorPoint{result.Config, 30000000, float64(result.SLO)}
	i++

	Predict(capPoints, CapacitorPoint{config: *configs[2], wkl: 20000000}, 100000)
}
