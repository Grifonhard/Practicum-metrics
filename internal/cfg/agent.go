package cfg

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v10"
)

type Agent struct {
	Addr           *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	Key            *string `env:"KEY"`
	RateLimit      *int    `env:"RATE_LIMIT"`
	CryptoKey      *string `env:"CRYPTO_KEY"`
	Config         *string `env:"CONFIG"`
}

type AgentFile struct {
	Address        *string `json:"address"`
	ReportInterval *int `json:"report_interval"`
	PollInterval   *int `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
	Key            *string `json:"key"`
	RateLimit      *int    `json:"rate_limit"`
}

type AgentFlags struct {
    Address        *string
	ReportInterval *int
	PollInterval   *int
	CryptoKey      *string
	Key            *string
	RateLimit      *int
    Config         *string
}

// Load загружает конфигурацию из разных источников
// все поля гарантированно будут не nil
// в случае отсутствия данных "" или 0
func (a *Agent) Load() error {
	err := env.Parse(a)
	if err != nil {
		return err
	}

    flags := &AgentFlags{}
    err = flags.loadAgentConfigFromFlags()
    if err != nil {
		return err
	}

    file := &AgentFile{}
    err = file.loadAgentConfigFromFile(a.Config, flags.Config)
    if err != nil {
		return err
	}

	return a.Resolve(flags, file)
}

func (a *Agent) Resolve(flags *AgentFlags, file *AgentFile) error {
    if a.Addr != nil {
    } else if flags.Address != nil && *flags.Address != "" {
        a.Addr = flags.Address
    } else if file.Address != nil {
        a.Addr = file.Address
    } else {
        addr := DEFAULTADDR
        a.Addr = &addr
    }
    if a.ReportInterval != nil {
    } else if flags.ReportInterval != nil && *flags.ReportInterval != 0 {
        a.ReportInterval = flags.ReportInterval
    } else if file.ReportInterval != nil {
        a.ReportInterval = file.ReportInterval
    } else {
        repInterval := DEFAULTREPORTINTERVAL
        a.ReportInterval = &repInterval
    }
    if a.PollInterval != nil {
    } else if flags.PollInterval != nil && *flags.PollInterval != 0 {
        a.PollInterval = flags.PollInterval
    } else if file.PollInterval != nil {
        a.PollInterval = file.PollInterval
    } else {
        pollInterval := DEFAULTPOLLINTERVAL
        a.PollInterval = &pollInterval
    }
    if a.Key != nil {
    } else if flags.Key != nil && *flags.Key != "" {
        a.Key = flags.Key
    } else if file.Key != nil {
        a.Key = file.Key
    } else {
        var key string
        a.Key = &key
    }
    if a.RateLimit != nil {
    } else if flags.RateLimit != nil && *flags.RateLimit != 0 {
        a.RateLimit = flags.RateLimit
    } else if file.RateLimit != nil {
        a.RateLimit = file.RateLimit
    } else {
        var rateLimit int
        a.RateLimit = &rateLimit
    }
    if a.CryptoKey != nil {
    } else if flags.CryptoKey != nil && *flags.CryptoKey != "" {
        a.CryptoKey = flags.CryptoKey
    } else if file.CryptoKey != nil {
        a.CryptoKey = file.CryptoKey
    } else {
        var cryptoKey string
        a.CryptoKey = &cryptoKey
    }
    return nil
}

func (a *AgentFlags) loadAgentConfigFromFlags() error {
    a.Address = flag.String("a", "", "адрес сервера")
	a.ReportInterval = flag.Int("r", 0, "секунд частота отправки метрик")
	a.PollInterval = flag.Int("p", DEFAULTPOLLINTERVAL, "секунд частота опроса метрик")
	a.Key = flag.String("k", "", "ключ для хэша")
	a.RateLimit = flag.Int("l", 0, "ограничение количества одновременно исходящих запросов")
	a.CryptoKey = flag.String("crypto-key", "", "path to RSA public key (for encryption)")
	a.Config = flag.String("c", "", "path to json config")

    flag.Parse()

    return nil
}

func (a *AgentFile) loadAgentConfigFromFile(pathEnv, pathFlag *string) error {
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
        return err
    }

    type interm struct {
        Address        *string `json:"address"`
        ReportInterval *string `json:"report_interval"`
        PollInterval   *string `json:"poll_interval"`
        CryptoKey      *string `json:"crypto_key"`
        Key            *string `json:"key"`
        RateLimit      *int    `json:"rate_limit"`
    }

    var im interm

    if err = json.Unmarshal(data, &im); err != nil {
        return err
    }

    a.Address = im.Address
    repInter, err := parseStrToInt(im.ReportInterval)
    if err != nil { return err }
    a.ReportInterval = repInter
    pollInter, err := parseStrToInt(im.PollInterval)
    if err != nil { return err }
    a.PollInterval = pollInter
    a.CryptoKey = im.CryptoKey
    a.Key = im.Key
    a.RateLimit = im.RateLimit

    return nil
}
