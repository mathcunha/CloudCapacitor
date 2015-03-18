package capacitor

import (
	"testing"
)

func TestMockExecuter(t *testing.T) {
	m := MockExecuter{"/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/wordpress_cpu_mem.csv", nil}
	err := m.Load()
	if err != nil {
		t.Errorf("config error")
	}
	t.Log(m)
}
