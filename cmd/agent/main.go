package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	cryptoutils "github.com/Grifonhard/Practicum-metrics/internal/crypto_utils"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	webclient "github.com/Grifonhard/Practicum-metrics/internal/web_client"
	"github.com/caarlos0/env/v10"
)

var (
	buildVersion = "NA"
	buildDate    = "NA"
	buildCommit  = "NA"
)

const (
	DEFAULTADDR           = "localhost:8080"
	DEFAULTREPORTINTERVAL = 10
	DEFAULTPOLLINTERVAL   = 2
	NA                    = "N/A"
)

type CFG struct {
	Addr           *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	Key            *string `env:"KEY"`
	RateLimit      *int    `env:"RATE_LIMIT"`
	CryptoKey      *string `env:"CRYPTO_KEY"`
}

func main() {
	address := flag.String("a", DEFAULTADDR, "адрес сервера")
	reportInterval := flag.Int("r", DEFAULTREPORTINTERVAL, "секунд частота отправки метрик")
	pollInterval := flag.Int("p", DEFAULTPOLLINTERVAL, "секунд частота опроса метрик")
	key := flag.String("k", "", "ключ для хэша")
	rateLimit := flag.Int("l", 0, "ограничение количества одновременно исходящих запросов")
	cryptoKeyPath := flag.String("crypto-key", "", "path to RSA public key (for encryption)")

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
	if cfg.RateLimit != nil {
		rateLimit = cfg.RateLimit
	}
	if cfg.CryptoKey != nil && *cfg.CryptoKey != "" {
		cryptoKeyPath = cfg.CryptoKey
	}

	if *cryptoKeyPath != "" {
		cryptoutils.PublicKey, err = cryptoutils.LoadPublicKey(*cryptoKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("public key successfully loaded")
	}

	generator := metgen.New()

	timerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	timerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)

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
			if *rateLimit == 0 {
				go webclient.SendMetric(fmt.Sprintf("http://%s/updates/", *address), generator, *key, webclient.SENDARRAY)
			} else {
				go webclient.SendMetricWithWorkerPool(fmt.Sprintf("http://%s/updates/", *address), generator, *key, *rateLimit)
			}
		}
	}
}

func showMeta() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Printf("Build date: %s\n", buildDate)
}
