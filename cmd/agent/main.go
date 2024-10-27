package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
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
	Addr           *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	Key			   *string `env:"KEY"`
}

func main() {
	address := flag.String("a", DEFAULTADDR, "адрес сервера")
	reportInterval := flag.Int("r", DEFAULTREPORTINTERVAL, "секунд частота отправки метрик")
	pollInterval := flag.Int("p", DEFAULTPOLLINTERVAL, "секунд частота опроса метрик")
	key := flag.String("k", "", "ключ для хэша")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Addr != nil {
		address = cfg.Addr
	}
	if cfg.PollInterval != nil {
		pollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval != nil {
		reportInterval = cfg.ReportInterval
	}
	if cfg.Key != nil {
		key = cfg.Key
	}

	generator := metgen.New()

	timerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	timerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

	err = logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-timerPoll.C:
			err = generator.Renew()
			if err != nil {
				logger.Error(fmt.Sprintf("Fail renew metrics: %s\n", err.Error()))
			}
		case <-timerReport.C:
			go webclient.SendMetric(fmt.Sprintf("http://%s/updates/", *address), generator, *key, webclient.SENDARRAY)
		}
	}
}
