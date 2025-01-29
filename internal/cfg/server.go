package cfg

type Server struct {
	Addr            *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDsn     *string `env:"DATABASE_DSN"`
	Key             *string `env:"KEY"`
	CryptoKey       *string `env:"CRYPTO_KEY"`
	Config         *string `env:"CONFIG"`
}

type ServerFlags struct {
	Address            *string
	StoreInterval   *int 
	FileStoragePath *string
	Restore         *bool
	DatabaseDsn     *string
	Key             *string
	CryptoKey       *string
	Config         *string
}

type ServerFile struct {
	Address       *string `json:"address"`
	StoreInterval *int `json:"store_interval"`
    Restore       bool   `json:"restore"`
    // store_interval в JSON файле может быть строкой "1s"
    // В Go флаге/env это int (секунды). Нужен парсинг
    
    // store_file -> это аналог FileStoragePath
    StoreFile    string `json:"store_file"`
    DatabaseDSN  string `json:"database_dsn"`
    CryptoKey    string `json:"crypto_key"`
}