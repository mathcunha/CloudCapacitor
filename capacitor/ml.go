package capacitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

type ML struct {
	capPoints             []CapacitorPoint
	uslThrput             *USL
	linearRegressionModel *LinearRegression
}

type Points []Point

type USL struct {
	Points                 Points
	Alpha, Beta, R2, N, Y1 float64
	Y1IsMax                bool
}

type LinearRegression struct {
	Points          Points
	Alpha, Beta, R2 float64
}

func init() {
	local = true
	data, err := ioutil.ReadFile("config/azureml.yml")
	if err != nil {
		log.Printf("error reading azure file: %v\n", err)
		data, err = ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/mathcunha/CloudCapacitor/config/azureml.yml")
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

func NewML(capPoints []CapacitorPoint, lr bool) (ml ML) {
	ml.capPoints = capPoints
	if lr {
		if points := pointsByThroughput(capPoints); len(points) > 1 {
			model := LinearRegression{Points: points}
			model.BuildModel()
			ml.linearRegressionModel = &model
		}
		return
	}
	if points := pointsByThroughput(capPoints); len(points) > 1 {
		usl := USL{Points: points}
		usl.BuildUSL()
		ml.uslThrput = &usl
	}
	return
}

func (ml *ML) Predict(capPoint CapacitorPoint) (performance float64, model string) {
	performance = -1
	model = ""

	if ml.linearRegressionModel != nil {
		performance = float64(capPoint.wkl) / ml.linearRegressionModel.Predict(workloadModelX(capPoint))
		model = "linearRegression"
		return
	}

	if points := pointsByConfiguration(ml.capPoints, capPoint); len(points) > 1 {
		usl := USL{Points: points}
		usl.BuildUSL()
		if isBetweenPoints(points, float64(capPoint.wkl)) {
			performance = usl.Predict(float64(capPoint.wkl))
			model = "uslByConfig"
			return
		}
	}
	/*
		if points := pointsByWorkload(capPoints, capPoint); len(points) > 1 {
			usl := USL{Points: points, Y1IsMax: true}
			usl.BuildUSL()
			if isBetweenPoints(points, workloadModelX(capPoint)) && usl.R2 >= 0.7 {
				performance = usl.Predict(workloadModelX(capPoint))
				model = "uslByWorkload"
				return
			}
		}
	*/
	if ml.uslThrput != nil && ml.uslThrput.R2 >= 0.7 {
		performance = float64(capPoint.wkl) / ml.uslThrput.Predict(workloadModelX(capPoint))
		model = "uslByThrput"
		return
	}
	return
}

func isBetweenPoints(points Points, x float64) bool {
	hasSmaller, hasHigher := false, false
	for _, v := range points {
		if v.X < x {
			hasSmaller = true
		}
		if v.X > x {
			hasHigher = true
		}
		if hasSmaller && hasHigher {
			break
		}
	}
	return hasSmaller && hasHigher
}

func pointsByThroughput(capPoints []CapacitorPoint) (points Points) {
	for _, v := range capPoints {
		points = append(points, Point{workloadModelX(v), float64(v.wkl) / v.performance})
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

func workloadModelX(v CapacitorPoint) (x float64) {
	x = (0.3 * float64(v.config.Size)) + (0.7 * float64(v.config.VM.CPU))
	return
}

func pointsByWorkload(capPoints []CapacitorPoint, capPoint CapacitorPoint) (points Points) {
	for _, v := range capPoints {
		if v.wkl == capPoint.wkl {
			points = append(points, Point{float64(workloadModelX(v)), float64(v.performance)})
		}
	}
	return
}

func (lr *LinearRegression) Predict(x float64) (y float64) {
	y = (x * lr.Alpha) + lr.Beta
	return
}

func (u *USL) Predict(x float64) (y float64) {
	kappa := u.Beta
	sigma := u.Alpha + kappa
	y = x / (1 + (sigma * (x - 1)) + (kappa * x * (x - 1)))
	y *= u.Y1
	return
}

func (lr *LinearRegression) BuildModel() {
	buf := bytes.NewBufferString("")
	json.NewEncoder(buf).Encode(lr.Points)

	cmd := exec.Command("Rscript", "--vanilla", os.Getenv("GOPATH")+"/src/github.com/mathcunha/CloudCapacitor/config/lr.R", buf.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		fmt.Printf("callRScript error calling Rscript :%v\n", err)
		return
	}
	if err := json.Unmarshal(out.Bytes(), &[]*LinearRegression{lr}); err != nil {
		log.Printf("callRScript error parsing response %v: %v\n", out.String(), err)
	}
	return
}

func (u *USL) callRScript(points string) {
	cmd := exec.Command("Rscript", "--vanilla", os.Getenv("GOPATH")+"/src/github.com/mathcunha/CloudCapacitor/config/usl.R", points)
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
			if len(values.Results.Output1.Value.Values) > 0 {
				usl := values.Results.Output1.Value.Values[0][0]
				usl = usl[1 : len(usl)-1]
				err = json.Unmarshal([]byte(usl), u)
				if err != nil {
					log.Printf("BuildUSL error parsing response %v: %v\n", usl, err)
				}
			} else {
				log.Printf("BuildUSL error parsing response %v\n", out)
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
		//factor := 1.0 / smaller.X
		if u.Y1IsMax {
			smaller.X, smaller.Y = 1, smaller.Y*smaller.X
		} else {
			smaller.X, smaller.Y = 1, smaller.Y/smaller.X
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
