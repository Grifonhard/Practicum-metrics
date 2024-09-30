package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	webclient "github.com/Grifonhard/Practicum-metrics/internal/web_client"
	"github.com/caarlos0/env/v10"
)

const (
	DEFAULTADDR           = "localhost:8080"
	DEFAULTREPORTINTERVAL = 10
	DEFAULTPOLLINTERVAL   = 2
)

type CFG struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	address := flag.String("a", DEFAULTADDR, "адрес сервера")
	reportInterval := flag.Int("r", DEFAULTREPORTINTERVAL, "секунд частота отправки метрик")
	pollInterval := flag.Int("p", DEFAULTPOLLINTERVAL, "секунд частота опроса метрик")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if _, exist := os.LookupEnv("ADDRESS"); exist {
		address = &cfg.Addr
	}
	if _, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		pollInterval = &cfg.PollInterval
	}
	if _, exist := os.LookupEnv("POLL_INTERVAL"); exist {
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
			go webclient.SendMetric(fmt.Sprintf("http://%s/update", *address), generator)
		}
	}
}
