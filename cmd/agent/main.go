package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	var cfg cfg.Agent
	err = cfg.Load()
	if err != nil {
		log.Fatal(err)
	}
	
	if *cfg.CryptoKey != "" {
		cryptoutils.PublicKey, err = cryptoutils.LoadPublicKey(*cfg.CryptoKey)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("public key successfully loaded")
	}

	generator := metgen.New()

	timerPoll := time.NewTicker(time.Duration(*cfg.PollInterval) * time.Second)
	defer timerPoll.Stop()
	timerReport := time.NewTicker(time.Duration(*cfg.ReportInterval) * time.Second)
	defer timerReport.Stop()

	showMeta()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	var wg sync.WaitGroup

	for {
		select {
		case <-timerPoll.C:
			err = generator.Renew()
			if err != nil {
				logger.Error(fmt.Sprintf("Fail renew metrics: %s\n", err.Error()))
			}
		case <-timerReport.C:
			if *cfg.RateLimit == 0 {
				go webclient.SendMetric(&wg, fmt.Sprintf("http://%s/updates/", *cfg.Addr), generator, *cfg.Key, webclient.SENDARRAY)
			} else {
				go webclient.SendMetricWithWorkerPool(&wg, fmt.Sprintf("http://%s/updates/", *cfg.Addr), generator, *cfg.Key, *cfg.RateLimit)
			}
		case <- ctx.Done():
			wg.Wait()
			logger.Info("agent shut down")
			return
		}
	}
}

func showMeta() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Printf("Build date: %s\n", buildDate)
}
