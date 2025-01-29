//go:generate ./generate_version.sh

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	cryptoutils "github.com/Grifonhard/Practicum-metrics/internal/crypto_utils"
	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/caarlos0/env/v10"
	"github.com/gin-gonic/gin"
)

var (
	buildVersion = "NA"
	buildDate    = "NA"
	buildCommit  = "NA"
)

type CFG struct {
	Addr            *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDsn     *string `env:"DATABASE_DSN"`
	Key             *string `env:"KEY"`
	CryptoKey       *string `env:"CRYPTO_KEY"`
}

func main() {
	addr := flag.String("a", DEFAULTADDR, "server address")
	storeInterval := flag.Int("i", DEFAULTSTOREINTERVAL, "backup interval")
	fileStoragePath := flag.String("f", "", "file storage path")
	restore := flag.Bool("r", DEFAULTRESTORE, "restore from backup")
	databaseDsn := flag.String("d", "", "database connect")
	key := flag.String("k", "", "ключ для хэша")
	cryptoKeyPath := flag.String("crypto-key", "", "Path to RSA private key (for decryption)")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Addr != nil {
		addr = cfg.Addr
	}
	if cfg.StoreInterval != nil {
		storeInterval = cfg.StoreInterval
	}
	if cfg.FileStoragePath != nil {
		fileStoragePath = cfg.FileStoragePath
	}
	if cfg.Restore != nil {
		restore = cfg.Restore
	}
	if cfg.DatabaseDsn != nil {
		databaseDsn = cfg.DatabaseDsn
	}
	if cfg.Key != nil {
		key = cfg.Key
	}
	if cfg.CryptoKey != nil && *cfg.CryptoKey != "" {
		cryptoKeyPath = cfg.CryptoKey
	}

	if *cryptoKeyPath != "" {
		cryptoutils.PrivateKey, err = cryptoutils.LoadPrivateKey(*cryptoKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("private key successfully loaded")
	}

	showMeta()

	err = logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	var dbInter psql.StorDB
	var db *psql.DB
	if *databaseDsn != "" {
		logger.Info(fmt.Sprintf("Database DSN: %s\n", *databaseDsn))
		db, err = psql.ConnectDB(*databaseDsn)
		if err != nil {
			log.Fatal(err)
		}
		if pingErr := db.Ping(); pingErr != nil {
			_ = db.Close()
			log.Fatal(err)
		}
		dbInter = db
	}

	stor, err := storage.New(*storeInterval, *fileStoragePath, *restore, dbInter)
	if err != nil {
		log.Fatal(err)
	}

	go stor.BackupLoop()

	r := initRouter(stor, db, *key)

	logger.Info(fmt.Sprintf("Server start %s\n", *addr))
	log.Fatal(r.Run(*addr))
}

func initRouter(stor *storage.MemStorage, db *psql.DB, key string) *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("../../templates/*")

	router.POST("/update/", web.ReqRespLogger(""), web.DataExtraction(), web.RespEncode(), web.Update(stor))
	router.POST("/update/:type/:name/:value", web.ReqRespLogger(""), web.DataExtraction(), web.Update(stor))
	router.POST("/updates/", web.PseudoAuth(key), cryptoutils.DecryptBody(), web.ReqRespLogger(key), web.DataExtraction(), web.Updates(stor))
	router.GET("/value/:type/:name", web.ReqRespLogger(""), web.DataExtraction(), web.Get(stor))
	router.POST("/value/", web.ReqRespLogger(""), web.RespEncode(), web.GetJSON(stor))
	router.GET("/", web.ReqRespLogger(""), web.RespEncode(), web.List(stor))
	if db != nil {
		router.GET("/ping", web.PingDB(db))
	}

	return router
}

func showMeta() {
	if buildVersion == "" {
		fmt.Printf("Build version: %s\n", NA)
	} else {
		fmt.Printf("Build version: %s\n", buildVersion)
	}
	if buildCommit == "" {
		fmt.Printf("Build commit: %s\n", NA)
	} else {
		fmt.Printf("Build commit: %s\n", buildCommit)
	}
	if buildDate == "" {
		fmt.Printf("Build date: %s\n", NA)
	} else {
		fmt.Printf("Build date: %s\n", buildDate)
	}
}
