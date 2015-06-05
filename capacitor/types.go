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
	Strict   int
}

type Configuration struct {
	Size int
	VM
}

func (c *Configuration) Price() float32 {
	return float32(c.Size) * c.VM.Price
}

func (c *Configuration) Mem() float32 {
	return float32(c.Size) * c.VM.Mem
}

func (c *Configuration) CPU() float32 {
	return float32(c.Size) * c.VM.CPU
}

func (c *Configuration) Strict() int {
	return c.VM.Strict
}

func (vm VM) String() string {
	return fmt.Sprintf("{category:%v, cpu:%v , mem:%v, price:%v, strict:%v}", vm.Category, vm.CPU, vm.Mem, vm.Price, vm.Strict)
}

func (c Configuration) String() string {
	return fmt.Sprintf("{vm:%v, size:%v, cpu:%v, mem:%v, price:%v}", c.VM.Name, c.Size, c.CPU(), c.Mem(), c.Price())
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

	//the array must be sorted by capacity
	for i, _ := range vms {
		vms[i].Strict = i + 1
	}

	return vms, nil
}
