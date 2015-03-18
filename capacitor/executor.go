package capacitor

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Result struct {
	Height int
	Config Configuration
	Performance
}

type Performance struct {
	Mem float32
	CPU float32
	SLO float32
}

type Executer interface {
	Execute(config Configuration, wkl string) (r Result)
}

type MockExecuter struct {
	FilePath string
	records  map[string][]string
}

func (e *MockExecuter) Load() error {
	file, err := os.Open(e.FilePath)
	if err != nil {
		log.Printf("error opening file %v:%v", e.FilePath, err)
		return err
	}
	records, err := csv.NewReader(file).ReadAll()

	if err != nil {
		log.Printf("error reading csv file %v:%v", e.FilePath, err)
	}

	mapa := make(map[string][]string)
	for _, v := range records {
		mapa[fmt.Sprintf("%v%v%v", v[0], v[1], v[2])] = []string{v[4], v[5], v[6]}
	}

	e.records = mapa
	return err
}

func (e *MockExecuter) Execute(config Configuration, wkl string) (result Result) {
	r := new(Result)

	r.Config = config

	key := fmt.Sprintf("%v%v%v", config.Size, config.VM.Name, wkl)

	perf, has := e.records[key]

	if has {
		p := Performance{}
		f32, _ := strconv.ParseFloat(perf[2], 32)
		p.Mem = float32(f32)

		f32, _ = strconv.ParseFloat(perf[1], 32)
		p.CPU = float32(f32)

		f32, _ = strconv.ParseFloat(perf[0], 32)
		p.SLO = float32(f32)
	} else {
		log.Printf("MockExecuter.Execute, key(%v) not found", key)
	}

	return *r
}
