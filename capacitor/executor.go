package capacitor

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	HighUsage = 90
	LowUsage  = 70
	HighDelta = 0.8
	LowDelta  = 0.6
)

type Result struct {
	Config Configuration
	Performance
}

type Performance struct {
	Mem float32
	CPU float32
	SLO float32
}

type Executor interface {
	Execute(config Configuration, wkl string) (r Result)
}

type MockExecutor struct {
	FilePath string
	Records  map[string][]string
}

func (e *MockExecutor) Load() error {
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

	e.Records = mapa
	return err
}

func (e MockExecutor) Execute(config Configuration, wkl string) (result Result) {
	r := new(Result)

	r.Config = config

	key := fmt.Sprintf("%v%v%v", config.Size, config.VM.Name, wkl)

	perf, has := e.Records[key]

	if has {
		p := Performance{}
		f32, _ := strconv.ParseFloat(perf[2], 32)
		p.Mem = float32(f32)

		f32, _ = strconv.ParseFloat(perf[1], 32)
		p.CPU = float32(f32)

		f32, _ = strconv.ParseFloat(perf[0], 32)
		p.SLO = float32(f32)

		r.Performance = p
	} else {
		log.Printf("MockExecutor.Execute, key(%v) not found", key)
	}

	return *r
}

func (result Result) String() (str string) {
	str = fmt.Sprintf("{Config:%v, Performance:%v}", result.Config, result.Performance)
	return
}

func (perf Performance) String() (str string) {
	str = fmt.Sprintf("{mem:%v, cpu:%v, slo:%v}", perf.Mem, perf.CPU, perf.SLO)
	return
}
