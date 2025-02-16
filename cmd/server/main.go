//go:generate ./generate_version.sh

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	var cfg cfg.Server
	err = cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	if *cfg.CryptoKey != "" {
		cryptoutils.PrivateKey, err = cryptoutils.LoadPrivateKey(*cfg.CryptoKey)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("private key successfully loaded")
	}

	var ipNet *net.IPNet
	if *cfg.TrustedSubnet != "" {
		_, ipNet, err = net.ParseCIDR(*cfg.TrustedSubnet)
		if err != nil {
			log.Fatalf("Ошибка парсинга trusted_subnet (%s): %v", cfg.TrustedSubnet, err)
		}
		log.Printf("Загруженная доверенная подсеть: %s", cfg.TrustedSubnet)
	}

	showMeta()

	var dbInter psql.StorDB
	var db *psql.DB
	if *cfg.DatabaseDsn != "" {
		logger.Info(fmt.Sprintf("Database DSN: %s\n", *cfg.DatabaseDsn))
		db, err = psql.ConnectDB(*cfg.DatabaseDsn)
		if err != nil {
			log.Fatal(err)
		}
		dbInter = db
	}

	stor, err := storage.New(*cfg.StoreInterval, *cfg.FileStoragePath, *cfg.Restore, dbInter)
	if err != nil {
		log.Fatal(err)
	}

	go stor.BackupLoop()

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	var wg sync.WaitGroup

	r := initRouter(&wg, stor, db, *cfg.Key, ipNet)
	
	logger.Info(fmt.Sprintf("Server start %s\n", *cfg.Addr))

	go func (){
		log.Fatal(r.Run(*cfg.Addr))
	}()
	
	<-sig
	wg.Wait()
	logger.Info("server shutdown")
}

func initRouter(wg *sync.WaitGroup, stor *storage.MemStorage, db *psql.DB, key string, ipNet *net.IPNet) *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("./templates/*")

	router.POST("/update/", web.WGadd(wg), web.ReqRespLogTScheck("", ipNet), web.DataExtraction(), web.RespEncode(), web.Update(wg, stor))
	router.POST("/update/:type/:name/:value", web.WGadd(wg), web.ReqRespLogTScheck("", ipNet), web.DataExtraction(), web.Update(wg, stor))
	router.POST("/updates/", web.WGadd(wg), web.PseudoAuth(key), cryptoutils.DecryptBody(), web.ReqRespLogTScheck(key, ipNet), web.DataExtraction(), web.Updates(wg, stor))
	router.GET("/value/:type/:name", web.ReqRespLogTScheck("", ipNet), web.DataExtraction(), web.Get(stor))
	router.POST("/value/", web.ReqRespLogTScheck("", ipNet), web.RespEncode(), web.GetJSON(stor))
	router.GET("/", web.ReqRespLogTScheck("", ipNet), web.RespEncode(), web.List(stor))
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
