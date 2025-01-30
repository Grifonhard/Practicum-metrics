//go:generate ./generate_version.sh

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Grifonhard/Practicum-metrics/internal/cfg"
	cryptoutils "github.com/Grifonhard/Practicum-metrics/internal/crypto_utils"
	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/gin-gonic/gin"
)

var (
	buildVersion = "NA"
	buildDate    = "NA"
	buildCommit  = "NA"
)

func main() {

	var cfg cfg.Server
	err := cfg.Load()

	if *cfg.CryptoKey != "" {
		cryptoutils.PrivateKey, err = cryptoutils.LoadPrivateKey(*cfg.CryptoKey)
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
	if *cfg.DatabaseDsn != "" {
		logger.Info(fmt.Sprintf("Database DSN: %s\n", *cfg.DatabaseDsn))
		db, err = psql.ConnectDB(*cfg.DatabaseDsn)
		if err != nil {
			log.Fatal(err)
		}
		if pingErr := db.Ping(); pingErr != nil {
			_ = db.Close()
			log.Fatal(err)
		}
		dbInter = db
	}

	stor, err := storage.New(*cfg.StoreInterval, *cfg.FileStoragePath, *cfg.Restore, dbInter)
	if err != nil {
		log.Fatal(err)
	}

	go stor.BackupLoop()

	r := initRouter(stor, db, *cfg.Key)

	logger.Info(fmt.Sprintf("Server start %s\n", *cfg.Addr))
	log.Fatal(r.Run(*cfg.Addr))
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
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Printf("Build date: %s\n", buildDate)
}
