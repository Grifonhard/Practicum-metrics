package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/cfg"
	cryptoutils "github.com/Grifonhard/Practicum-metrics/internal/crypto_utils"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	webclient "github.com/Grifonhard/Practicum-metrics/internal/web_client"
)

var (
	buildVersion = "NA"
	buildDate    = "NA"
	buildCommit  = "NA"
)

func main() {
	
	var cfg cfg.Agent
	err := cfg.Load()

	if *cfg.CryptoKey != "" {
		cryptoutils.PublicKey, err = cryptoutils.LoadPublicKey(*cfg.CryptoKey)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("public key successfully loaded")
	}

	generator := metgen.New()

	timerPoll := time.NewTicker(time.Duration(*cfg.PollInterval) * time.Second)
	timerReport := time.NewTicker(time.Duration(*cfg.ReportInterval) * time.Second)

	err = logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	showMeta()

	for {
		select {
		case <-timerPoll.C:
			err = generator.Renew()
			if err != nil {
				logger.Error(fmt.Sprintf("Fail renew metrics: %s\n", err.Error()))
			}
		case <-timerReport.C:
			if *cfg.RateLimit == 0 {
				go webclient.SendMetric(fmt.Sprintf("http://%s/updates/", *cfg.Addr), generator, *cfg.Key, webclient.SENDARRAY)
			} else {
				go webclient.SendMetricWithWorkerPool(fmt.Sprintf("http://%s/updates/", *cfg.Addr), generator, *cfg.Key, *cfg.RateLimit)
			}
		}
	}
}

func showMeta() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Printf("Build date: %s\n", buildDate)
}
