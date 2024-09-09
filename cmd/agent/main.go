package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	"github.com/caarlos0/env/v10"
)

const (
	DEFAULT_ADDR = "localhost:8080"
	DEFAULT_REPORT_INTERVAL = 10
	DEFAULT_POLL_INTERVAL = 2
)

type CFG struct{
	Addr	string `env:"ADDRESS"`
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval int `env:"POLL_INTERVAL"`
}

func main() {
	var address string
	reportInterval := flag.Int("r", DEFAULT_REPORT_INTERVAL, "секунд частота отправки метрик")
	pollInterval := flag.Int("p", DEFAULT_POLL_INTERVAL, "секунд частота опроса метрик")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil{
		log.Fatal(err)
	}

	if cfg.Addr != ""{
		address = cfg.Addr
	} else{
		address = DEFAULT_ADDR
	}
	if cfg.PollInterval != 0 {
		pollInterval = &cfg.PollInterval
	}
	if cfg.ReportInterval != 0 {
		reportInterval = &cfg.ReportInterval
	}

	generator := metgen.New()

	timerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	timerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

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
					sendMetric(address, "gauge", k, v)
				}
				for k, v := range counter {
					sendMetric(address, "counter", k, v)
				}
				return
			}()
		}
	}
}

func sendMetric(addr, mType, mKey, mValue string) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", addr, mType, mKey, mValue)

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
