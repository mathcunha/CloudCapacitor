package capacitor

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type VM struct {
	CPU      float32
	Mem      float32
	Price    float32
	Category string
	Name     string
}

type Configuration struct {
	Size int
	VM
}

func (c *Configuration) Price() float32 {
	return float32(c.Size) * c.VM.Price
}

func (vm VM) String() string {
	return fmt.Sprintf("{category:%v, cpu:%v , mem:%v, price:%v}", vm.Category, vm.CPU, vm.Mem, vm.Price)
}

func (c Configuration) String() string {
	return fmt.Sprintf("{vm:%v, size:%v}", c.VM, c.Size)
}

func LoadTypes(path string) (vms []VM, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("error reading file %v:%v", path, err)
		return nil, err
	}

	vms = []VM{}
	err = yaml.Unmarshal([]byte(data), &vms)
	if err != nil {
		log.Printf("error parsing file %v:%v", path, err)
		return nil, err
	}

	return vms, nil
}
