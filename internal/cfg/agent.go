package cfg

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/caarlos0/env/v10"
)

type Agent struct {
	Addr           *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	Key            *string `env:"KEY"`
	RateLimit      *int    `env:"RATE_LIMIT"`
	CryptoKey      *string `env:"CRYPTO_KEY"`
	TrustedSubnet  *string `env:"TRUSTED_SUBNET"`
	Config         *string `env:"CONFIG"`
}

type AgentFile struct {
	Address        *string `json:"address"`
	ReportInterval *int    `json:"report_interval"`
	PollInterval   *int    `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
	Key            *string `json:"key"`
	RateLimit      *int    `json:"rate_limit"`
	TrustedSubnet  *string `json:"trusted_subnet"`
}

type AgentFlags struct {
	Address        *string
	ReportInterval *int
	PollInterval   *int
	CryptoKey      *string
	Key            *string
	RateLimit      *int
	TrustedSubnet  *string
	Config         *string
}

// Load загружает конфигурацию из разных источников
// все поля гарантированно будут не nil
// в случае отсутствия данных "" или 0
func (a *Agent) Load() error {
	// caarlos0/env криво парсит структуры с указателями
	type agWhithoutPtr struct {
		Addr           string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
		Key            string `env:"KEY"`
		RateLimit      int    `env:"RATE_LIMIT"`
		CryptoKey      string `env:"CRYPTO_KEY"`
		TrustedSubnet  string `env:"TRUSTED_SUBNET"`
		Config         string `env:"CONFIG"`
	}

	var a2 agWhithoutPtr

	err := env.Parse(&a2)
	if err != nil {
		return err
	}

	a.Addr = &a2.Addr
	a.ReportInterval = &a2.ReportInterval
	a.PollInterval = &a2.PollInterval
	a.Key = &a2.Key
	a.RateLimit = &a2.RateLimit
	a.CryptoKey = &a2.CryptoKey
	a.TrustedSubnet = &a2.TrustedSubnet
	a.Config = &a2.Config

	flags := &AgentFlags{}
	err = flags.loadConfigFromFlags()
	if err != nil {
		return err
	}

	file := &AgentFile{}
	err = file.loadConfigFromFile(a.Config, flags.Config)
	if errors.Is(err, ErrCFGFile) {
		logger.Info(err.Error())
	} else if err != nil {
		return err
	}

	return a.Resolve(flags, file)
}

func (a *Agent) Resolve(flags *AgentFlags, file *AgentFile) error {
	if a.Addr != nil && *a.Addr != "" {
	} else if flags.Address != nil && *flags.Address != "" {
		a.Addr = flags.Address
	} else if file.Address != nil {
		a.Addr = file.Address
	} else {
		addr := DEFAULTADDR
		a.Addr = &addr
	}
	if a.ReportInterval != nil && *a.ReportInterval != 0 {
	} else if flags.ReportInterval != nil && *flags.ReportInterval != 0 {
		a.ReportInterval = flags.ReportInterval
	} else if file.ReportInterval != nil {
		a.ReportInterval = file.ReportInterval
	} else {
		repInterval := DEFAULTREPORTINTERVAL
		a.ReportInterval = &repInterval
	}
	if a.PollInterval != nil && *a.PollInterval != 0 {
	} else if flags.PollInterval != nil && *flags.PollInterval != 0 {
		a.PollInterval = flags.PollInterval
	} else if file.PollInterval != nil {
		a.PollInterval = file.PollInterval
	} else {
		pollInterval := DEFAULTPOLLINTERVAL
		a.PollInterval = &pollInterval
	}
	if a.Key != nil && *a.Key != "" {
	} else if flags.Key != nil && *flags.Key != "" {
		a.Key = flags.Key
	} else if file.Key != nil {
		a.Key = file.Key
	} else {
		var key string
		a.Key = &key
	}
	if a.RateLimit != nil && *a.RateLimit != 0 {
	} else if flags.RateLimit != nil && *flags.RateLimit != 0 {
		a.RateLimit = flags.RateLimit
	} else if file.RateLimit != nil {
		a.RateLimit = file.RateLimit
	} else {
		var rateLimit int
		a.RateLimit = &rateLimit
	}
	if a.CryptoKey != nil && *a.CryptoKey != "" {
	} else if flags.CryptoKey != nil && *flags.CryptoKey != "" {
		a.CryptoKey = flags.CryptoKey
	} else if file.CryptoKey != nil {
		a.CryptoKey = file.CryptoKey
	} else {
		var cryptoKey string
		a.CryptoKey = &cryptoKey
	}
	if a.TrustedSubnet != nil && *a.TrustedSubnet != "" {
	} else if flags.TrustedSubnet != nil && *flags.TrustedSubnet != "" {
		a.TrustedSubnet = flags.TrustedSubnet
	} else if file.TrustedSubnet != nil {
		a.TrustedSubnet = file.TrustedSubnet
	} else {
		var ts string
		a.TrustedSubnet = &ts
	}
	return nil
}

func (a *AgentFlags) loadConfigFromFlags() error {
	a.Address = flag.String("a", "", "адрес сервера")
	a.ReportInterval = flag.Int("r", 0, "секунд частота отправки метрик")
	a.PollInterval = flag.Int("p", 0, "секунд частота опроса метрик")
	a.Key = flag.String("k", "", "ключ для хэша")
	a.RateLimit = flag.Int("l", 0, "ограничение количества одновременно исходящих запросов")
	a.CryptoKey = flag.String("crypto-key", "", "path to RSA public key (for encryption)")
	a.TrustedSubnet = flag.String("-t", "", "trusted subnet")
	a.Config = flag.String("c", "", "path to json config")

	flag.Parse()

	return nil
}

func (a *AgentFile) loadConfigFromFile(pathEnv, pathFlag *string) error {
	var path string
	if pathEnv != nil {
		path = *pathEnv
	} else if pathFlag != nil {
		path = *pathFlag
	} else {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w %s", ErrCFGFile, err.Error())
	}

	type interm struct {
		Address        *string `json:"address"`
		ReportInterval *string `json:"report_interval"`
		PollInterval   *string `json:"poll_interval"`
		CryptoKey      *string `json:"crypto_key"`
		Key            *string `json:"key"`
		RateLimit      *int    `json:"rate_limit"`
		TrustedSubnet  *string `json:"trusted_subnet"`
	}

	var im interm

	if err = json.Unmarshal(data, &im); err != nil {
		return err
	}

	a.Address = im.Address
	repInter, err := parseStrToInt(im.ReportInterval)
	if err != nil {
		return err
	}
	a.ReportInterval = repInter
	pollInter, err := parseStrToInt(im.PollInterval)
	if err != nil {
		return err
	}
	a.PollInterval = pollInter
	a.CryptoKey = im.CryptoKey
	a.Key = im.Key
	a.RateLimit = im.RateLimit
	a.TrustedSubnet = im.TrustedSubnet

	return nil
}
