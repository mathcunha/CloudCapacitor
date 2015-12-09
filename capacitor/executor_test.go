package capacitor

import (
	"testing"
)

func TestMockExecutor(t *testing.T) {
	m := MockExecutor{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err := m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	result := m.Execute(Configuration{1, VM{1.0, 1.0, 1.0, "m3", "m3_medium", 1}, ""}, "1000")
	t.Logf("%v\n", result)
}
