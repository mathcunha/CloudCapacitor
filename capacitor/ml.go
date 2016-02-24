package capacitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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
	Points                 Points
	Alpha, Beta, R2, N, Y1 float64
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
}

func Predict(capPoints []CapacitorPoint, capPoint CapacitorPoint, slo float64) (performance float64) {
	uslByWorkload := USL{Points: pointsByWorkload(capPoints, capPoint)}
	uslByWorkload.BuildUSL()
	fmt.Println(uslByWorkload)
	performance = uslByWorkload.Predict(float64(capPoint.config.CPU()))
	fmt.Printf("prediction of %v is %f", capPoint, performance)
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

func (u *USL) Predict(x float64) (y float64) {
	kappa := u.Beta
	sigma := u.Alpha + kappa
	y = x / (1 + (sigma * (x - 1)) + (kappa * x * (x - 1)))
	y *= u.Y1
	return
}

func (u *USL) BuildUSL() {
	u.R2 = -1
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
	u.Y1 = points[0].Y

	buf := bytes.NewBufferString("")
	json.NewEncoder(buf).Encode(points)

	body := fmt.Sprintf("{\"Inputs\": {\"input1\": {\"ColumnNames\": [\"points\"], \"Values\": [[%q]]}}, \"GlobalParameters\": {}}", buf.String())
	out, err := CallAzureMLService(body, azureAuth, azureEndPoint)
	if err != nil {
		log.Printf("BuildUSL error calling AzureMLService: %v\n", err)
	} else {
		values := struct {
			Results struct {
				Output1 struct {
					//Value map[string][][]string
					Value struct {
						Values [][]string
					}
				} `json:"output1"`
			}
		}{}

		err = json.Unmarshal([]byte(out), &values)
		if err != nil {
			log.Printf("BuildUSL error parsing response %v: %v\n", out, err)
		} else {
			usl := values.Results.Output1.Value.Values[0][0]
			usl = usl[1 : len(usl)-1]
			err = json.Unmarshal([]byte(usl), u)
			if err != nil {
				log.Printf("BuildUSL error parsing response %v: %v\n", usl, err)
			}
		}
	}
}

func CallAzureMLService(body, auth, endpoint string) (out string, err error) {
	fmt.Printf("calling AzureMLService, endpoint:%s, auth:%s, body:%s\n", endpoint, auth, body)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	cli := &http.Client{Timeout: 2 * time.Minute}
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/json")
	res, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(resBody), nil
}
