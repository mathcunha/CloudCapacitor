package capacitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

var azureEndPoint string
var azureAuth string
var local bool

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
	Y1IsMax                bool
}

func init() {
	local = true
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
	local = false
}

func Predict(capPoints []CapacitorPoint, capPoint CapacitorPoint) (performance float64) {
	performance = -1
	if points := pointsByThroughput(capPoints, capPoint); len(points) > 1 {
		usl := USL{Points: points}
		usl.BuildUSL()
		if usl.R2 >= 0.7 {
			performance = float64(capPoint.wkl) / usl.Predict(float64(capPoint.config.CPU()))
			fmt.Printf("uslByThroughput prediction of %v is %f\n", capPoint, performance)
		}
	}
	if points := pointsByWorkload(capPoints, capPoint); len(points) > 1 {
		usl := USL{Points: points, Y1IsMax: true}
		usl.BuildUSL()
		if usl.R2 >= 0.7 {
			performance = usl.Predict(float64(capPoint.config.CPU()))
			fmt.Printf("uslByWorkload prediction of %v is %f\n", capPoint, performance)
		}
	}

	if points := pointsByConfiguration(capPoints, capPoint); len(points) > 1 {
		usl := USL{Points: points}
		usl.BuildUSL()
		performance = usl.Predict(float64(capPoint.wkl))
		fmt.Printf("uslByConfiguration prediction of %v is %f\n", capPoint, performance)
	}
	return
}

func pointsByThroughput(capPoints []CapacitorPoint, capPoint CapacitorPoint) (points Points) {
	for _, v := range capPoints {
		points = append(points, Point{float64(v.config.CPU()), float64(v.wkl) / v.performance})
	}
	return
}

func pointsByConfiguration(capPoints []CapacitorPoint, capPoint CapacitorPoint) (points Points) {
	for _, v := range capPoints {
		if v.config.Size == capPoint.config.Size && v.config.VM.Name == capPoint.config.VM.Name {
			points = append(points, Point{float64(v.wkl), v.performance})
		}
	}
	return
}

func pointsByWorkload(capPoints []CapacitorPoint, capPoint CapacitorPoint) (points Points) {
	for _, v := range capPoints {
		if v.wkl == capPoint.wkl {
			points = append(points, Point{float64(v.config.CPU()), float64(v.performance)})
		}
	}
	return
}

func (u *USL) Predict(x float64) (y float64) {
	kappa := u.Beta
	sigma := u.Alpha + kappa
	y = x / (1 + (sigma * (x - 1)) + (kappa * x * (x - 1)))
	y *= u.Y1
	return
}

func (u *USL) callRScript(points string) {
	cmd := exec.Command("Rscript", "--vanilla", "/home/vagrant/go/src/github.com/mathcunha/CloudCapacitor/config/usl.R", points)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		fmt.Printf("callRScript error calling Rscript :%v\n", err)
		return
	}
	if err := json.Unmarshal(out.Bytes(), &[]*USL{u}); err != nil {
		log.Printf("callRScript error parsing response %v: %v\n", out.String(), err)
	}
	return
}

func (u *USL) callAzureML(points string) {
	body := fmt.Sprintf("{\"Inputs\": {\"input1\": {\"ColumnNames\": [\"points\"], \"Values\": [[%q]]}}, \"GlobalParameters\": {}}", points)
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
		factor := 1.0 / smaller.X
		if u.Y1IsMax {
			smaller.X, smaller.Y = 1, smaller.Y*(1+factor)
		} else {
			smaller.X, smaller.Y = 1, smaller.Y*factor
		}
		points = append(points, smaller)
	}

	points = append(points, u.Points...)
	u.Y1 = points[0].Y

	buf := bytes.NewBufferString("")
	json.NewEncoder(buf).Encode(points)
	if local {
		u.callRScript(buf.String())
	} else {
		u.callAzureML(buf.String())
	}
}

func CallAzureMLService(body, auth, endpoint string) (out string, err error) {
	//fmt.Printf("calling AzureMLService, endpoint:%s, auth:%s, body:%s\n", endpoint, auth, body)
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
