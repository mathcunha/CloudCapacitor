package capacitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var azureEndPoint string
var azureAuth string

type CapacitorPoint struct {
	config      Configuration
	wkl         int
	performance float64
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Points []Point

type USL struct {
	Points               Points
	Alpha, Beta, R2, Max float64
}

func init() {
	data, err := ioutil.ReadFile("config/azureml.yml")
	if err != nil {
		log.Printf("error reading azure file: %v\n", err)
		data, err = ioutil.ReadFile("/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/azureml.yml")
		if err != nil {
			log.Printf("error reading azure file: %v\n", err)
			return
		}
	}

	config := struct {
		Auth     string
		Endpoint string
	}{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Printf("error parsing azure file:%v\n", err)
		return
	}
	azureEndPoint = config.Endpoint
	azureAuth = config.Auth
	fmt.Printf("%v\n", config)
}

func Predict(capPoints []CapacitorPoint, capPoint CapacitorPoint, slo float64) (performance float64) {
	uslByWorkload := USL{Points: pointsByWorkload(capPoints, capPoint)}
	uslByWorkload.BuildUSL()
	return
}

func pointsByWorkload(capPoints []CapacitorPoint, capPoint CapacitorPoint) (points Points) {
	for _, v := range capPoints {
		if v.wkl == capPoint.wkl {
			fmt.Printf("# of vCPUS %f for config %v\n", v.config.CPU(), v.config)
			points = append(points, Point{float64(v.config.CPU()), float64(v.performance)})
		}
	}
	fmt.Printf("original capPoints has %d elements and wokload filtered by %d has %d \n", len(capPoints), capPoint.wkl, len(points))
	return
}

func (u *USL) BuildUSL() {
	smaller := u.Points[0]
	for _, v := range u.Points {
		if v.X < smaller.X {
			smaller = v
		}
	}

	points := Points{}
	if smaller.X != 1 {
		smaller.X, smaller.Y = 1, smaller.Y/smaller.X
		points = append(points, smaller)
	}

	points = append(points, u.Points...)

	fmt.Printf("USL model has %d points, but model has %d\n", len(u.Points), len(points))

	buf := bytes.NewBufferString("")
	json.NewEncoder(buf).Encode(points)

	body := fmt.Sprintf("{\"Inputs\": {\"input1\": {\"ColumnNames\": [\"points\"], \"Values\": [[%q]]}}, \"GlobalParameters\": {}}", buf.String())
	fmt.Println(body)
}
