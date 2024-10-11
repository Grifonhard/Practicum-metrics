package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/caarlos0/env/v10"
	"github.com/gin-gonic/gin"
)

const (
	DEFAULTADDR          = "localhost:8080"
	DEFAULTSTOREINTERVAL = 300
	DEFAULTRESTORE       = true
)

type CFG struct {
	Addr            *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDsn		*string `env:"DATABASE_DSN"`
}

func main() {
	addr := flag.String("a", DEFAULTADDR, "server address")
	storeInterval := flag.Int("i", DEFAULTSTOREINTERVAL, "backup interval")
	fileStoragePath := flag.String("f", "", "file storage path")
	restore := flag.Bool("r", DEFAULTRESTORE, "restore from backup")
	databaseDsn := flag.String("d","","database connect")

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

	err = logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	stor, err := storage.New(*storeInterval, *fileStoragePath, *restore)
	if err != nil {
		log.Fatal(err)
	}

	var db *psql.DB
	if *databaseDsn != "" {
		db, err = psql.ConnectDB(*databaseDsn)
	}

	go stor.BackupLoop()

	r := initRouter(stor, db)

	fmt.Printf("Server start %s\n", *addr)
	log.Fatal(r.Run(*addr))
}

func initRouter(stor *storage.MemStorage, db *psql.DB) *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.POST("/update", web.ReqRespLogger(), web.DataExtraction(), web.RespEncode(), web.Update(stor))
	router.POST("/update/:type/:name/:value", web.ReqRespLogger(), web.DataExtraction(), web.Update(stor))
	router.GET("/value/:type/:name", web.ReqRespLogger(), web.DataExtraction(), web.Get(stor))
	router.POST("/value/", web.ReqRespLogger(), web.RespEncode(), web.GetJSON(stor))
	router.GET("/", web.ReqRespLogger(), web.RespEncode(), web.List(stor))
	if db != nil {
		router.GET("/ping", web.PingDB(db))
	}

	return router
}
