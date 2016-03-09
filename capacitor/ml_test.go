package capacitor

import (
	"fmt"
	"testing"
)

func TestRScript(t *testing.T) {
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

	points := pointsByWorkload(capPoints, CapacitorPoint{config: *configs[1], wkl: 10000000})

	usl := USL{Points: points}
	usl.BuildUSL()
}

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
	capPoints := make([]CapacitorPoint, 2, 2)

	result := m.Execute(*configs[0], "10000000")
	capPoints[i] = CapacitorPoint{result.Config, 10000000, float64(result.SLO)}
	i++

	result = m.Execute(*configs[len(configs)-1], "50000000")
	capPoints[i] = CapacitorPoint{result.Config, 50000000, float64(result.SLO)}
	i++

	ml := NewML(capPoints)
	prediction, modelName := ml.Predict(CapacitorPoint{config: *configs[1], wkl: 20000000})
	result = m.Execute(*configs[1], "20000000")
	fmt.Printf("%q,%.4f,%.4f\n", modelName, result.SLO, prediction)

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

	result = m.Execute(*configs[0], "40000000")
	capPoints[i] = CapacitorPoint{result.Config, 40000000, float64(result.SLO)}
	i++

	ml := NewML(capPoints)
	prediction, modelName := ml.Predict(CapacitorPoint{config: *configs[0], wkl: 40000000})
	result = m.Execute(*configs[0], "40000000")
	fmt.Printf("%q,%.4f,%.4f\n", modelName, result.SLO, prediction)
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

	ml := NewML(capPoints)
	prediction, modelName := ml.Predict(CapacitorPoint{config: *configs[2], wkl: 20000000})
	result = m.Execute(*configs[2], "20000000")
	fmt.Printf("%q,%.4f,%.4f\n", modelName, result.SLO, prediction)
}

func TestMachineLearningConfigurationModels(t *testing.T) {
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
	wkls := []int{10000000, 20000000, 30000000, 40000000, 50000000}
	configs := (*dspace.configs)["c3"]
	i := 0
	capPoints := make([]CapacitorPoint, len(configs)*len(wkls), len(configs)*len(wkls))
	for _, wkl := range wkls {
		for _, c := range configs {
			result := m.Execute(*c, fmt.Sprintf("%d", wkl))
			capPoints[i] = CapacitorPoint{result.Config, wkl, float64(result.SLO)}
			i++
		}
	}

	result := m.Execute(*configs[2], "20000000")
	ml := NewML(capPoints)
	ml.Predict(CapacitorPoint{config: *configs[2], wkl: 20000000})
	t.Logf("usl real is %f\n", result.SLO)

	result = m.Execute(*configs[8], "30000000")
	ml.Predict(CapacitorPoint{config: *configs[2], wkl: 30000000})
	t.Logf("usl real is %f\n", result.SLO)
}
