package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/met_gen"
)

const (
	ADDR           = "localhost"
	PORT           = "8080"
	POLLINTERVAL   = 2
	REPORTINTERVAL = 10
)

func main() {
	generator := metgen.New()

	timerPoll := time.NewTicker(time.Duration(POLLINTERVAL) * time.Second)
	timerReport := time.NewTicker(time.Duration(REPORTINTERVAL) * time.Second)

	var err error

	for {
		select {
		case <-timerPoll.C:
			err = generator.Renew()
			if err != nil {
				fmt.Printf("Fail renew metrics: %s", err.Error())
			}
		case <-timerReport.C:
			gauge, counter, err := generator.Collect()
			if err != nil {
				fmt.Printf("Fail collect metrics: %s", err.Error())
			}
			go func() {
				for k, v := range gauge {
					sendMetric("gauge", k, v)
				}
				for k, v := range counter {
					sendMetric("counter", k, v)
				}
				return
			}()
		}
	}
}

func sendMetric(mType, mKey, mValue string) {
	url := fmt.Sprintf("http://%s:%s/update/%s/%s/%s", ADDR, PORT, mType, mKey, mValue)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Printf("Fail while create request: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	cl := &http.Client{}

	resp, err := cl.Do(req)
	if err != nil {
		fmt.Printf("Fail while sending metrics: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	fmt.Println("success send")
}
