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

type Server struct {
	Addr            *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDsn     *string `env:"DATABASE_DSN"`
	Key             *string `env:"KEY"`
	CryptoKey       *string `env:"CRYPTO_KEY"`
	Config          *string `env:"CONFIG"`
}

type ServerFlags struct {
	Address         *string
	StoreInterval   *int
	FileStoragePath *string
	Restore         *bool
	DatabaseDsn     *string
	Key             *string
	CryptoKey       *string
	Config          *string
}

type ServerFile struct {
	Address       *string `json:"address"`
	StoreInterval *int    `json:"store_interval"`
	Restore       *bool   `json:"restore"`
	StoreFile     *string `json:"store_file"`
	DatabaseDSN   *string `json:"database_dsn"`
	Key           *string `json:"key"`
	CryptoKey     *string `json:"crypto_key"`
}

// Load загружает конфигурацию из разных источников
// все поля гарантированно будут не nil
// в случае отсутствия данных "" или 0
func (s *Server) Load() error {
	err := env.Parse(s)
	if err != nil {
		return err
	}

	flags := &ServerFlags{}
	err = flags.loadConfigFromFlags()
	if err != nil {
		return err
	}

	file := &ServerFile{}
	err = file.loadConfigFromFile(s.Config, flags.Config)
	if errors.Is(err, ErrCFGFile) {
        logger.Info(err.Error())
    } else if err != nil {
		return err
	}

	return s.Resolve(flags, file)
}

func (s *Server) Resolve(flags *ServerFlags, file *ServerFile) error {
	if s.Addr != nil {
	} else if flags.Address != nil && *flags.Address != "" {
		s.Addr = flags.Address
	} else if file.Address != nil {
		s.Addr = file.Address
	} else {
		addr := DEFAULTADDR
		s.Addr = &addr
	}
	if s.StoreInterval != nil {
	} else if flags.StoreInterval != nil && *flags.StoreInterval != 0 {
		s.StoreInterval = flags.StoreInterval
	} else if file.StoreInterval != nil {
		s.StoreInterval = file.StoreInterval
	} else {
		interval := DEFAULTSTOREINTERVAL
		s.StoreInterval = &interval
	}
	if s.Restore != nil {
	} else if flags.Restore != nil {
		s.Restore = flags.Restore
	} else if file.Restore != nil {
		s.Restore = file.Restore
	} else {
		restore := DEFAULTRESTORE
		s.Restore = &restore
	}
	if s.FileStoragePath != nil {
	} else if flags.FileStoragePath != nil && *flags.FileStoragePath != "" {
		s.FileStoragePath = flags.FileStoragePath
	} else if file.StoreFile != nil {
		s.FileStoragePath = file.StoreFile
	} else {
		var file string
		s.FileStoragePath = &file
	}
	if s.DatabaseDsn != nil {
	} else if flags.DatabaseDsn != nil && *flags.DatabaseDsn != "" {
		s.DatabaseDsn = flags.DatabaseDsn
	} else if file.DatabaseDSN != nil {
		s.DatabaseDsn = file.DatabaseDSN
	} else  {
		var dbDSN string
		s.DatabaseDsn = &dbDSN
	}
	if s.Key != nil {
	} else if flags.Key != nil && *flags.Key != "" {
		s.Key = flags.Key
	} else if file.Key != nil {
		s.Key = file.Key
	} else {
		var key string
		s.Key = &key
	}
	if s.CryptoKey != nil {
	} else if flags.CryptoKey != nil && *flags.CryptoKey != "" {
		s.CryptoKey = flags.CryptoKey
	} else if file.CryptoKey != nil {
		s.CryptoKey = file.CryptoKey
	} else {
		var cryptoKey string
		s.CryptoKey = &cryptoKey
	}
	return nil
}

func (s *ServerFlags) loadConfigFromFlags() error {
	s.Address = flag.String("a", "", "server address")
	s.StoreInterval = flag.Int("i", 0, "backup interval")
	s.FileStoragePath = flag.String("f", "", "file storage path")
	s.Restore = flag.Bool("r", false, "restore from backup")
	s.DatabaseDsn = flag.String("d", "", "database connect")
	s.Key = flag.String("k", "", "ключ для хэша")
	s.CryptoKey = flag.String("crypto-key", "", "Path to RSA private key (for decryption)")
	s.Config = flag.String("c", "", "path to json config")

	flag.Parse()

	return nil
}

func (s *ServerFile) loadConfigFromFile(pathEnv, pathFlag *string) error {

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
		Address       *string `json:"address"`
		StoreInterval *string `json:"store_interval"`
		Restore       *bool   `json:"restore"`
		StoreFile     *string `json:"store_file"`
		DatabaseDSN   *string `json:"database_dsn"`
		Key           *string `json:"key"`
		CryptoKey     *string `json:"crypto_key"`
	}

	var im interm

	if err = json.Unmarshal(data, &im); err != nil {
		return err
	}

	storInterval, err := parseStrToInt(im.StoreInterval)
	if err != nil {
		return err
	}

	s.Address = im.Address
	s.StoreInterval = storInterval
	s.Restore = im.Restore
	s.StoreFile = im.StoreFile
	s.DatabaseDSN = im.DatabaseDSN
	s.Key = im.Key
	s.CryptoKey = im.CryptoKey

	return nil
}
