package main

import (
	"flag"
	"fmt"
	"log"

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
	Addr            string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE" envDefault:"true"`
}

func main() {
	addr := flag.String("a", DEFAULTADDR, "server port")
	storeInterval := flag.Int("i", DEFAULTSTOREINTERVAL, "backup interval")
	fileStoragePath := flag.String("f", "", "file storage path")
	restore := flag.Bool("r", DEFAULTRESTORE, "restore from backup")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Addr != "" {
		*addr = cfg.Addr
	}
	if cfg.StoreInterval != 0 {
		*storeInterval = cfg.StoreInterval
	}
	if cfg.FileStoragePath != "" {
		*fileStoragePath = cfg.FileStoragePath
	}
	if !cfg.Restore {
		*restore = cfg.Restore
	}

	err = logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	stor, err := storage.New(*storeInterval, *fileStoragePath, *restore)
	if err != nil {
		panic(err)
	}

	go stor.BackupLoop()

	r := initRouter()

	r.POST("/update", web.ReqRespLogger(), web.DataExtraction(), web.RespEncode(), web.Update(stor))
	r.POST("/update/:type/:name/:value", web.ReqRespLogger(), web.DataExtraction(), web.Update(stor))
	r.GET("/value/:type/:name", web.ReqRespLogger(), web.DataExtraction(), web.Get(stor))
	r.POST("/value/", web.ReqRespLogger(), web.RespEncode(), web.GetJSON(stor))
	r.GET("/", web.ReqRespLogger(), web.RespEncode(), web.List(stor))

	fmt.Printf("Server start %s\n", *addr)
	log.Fatal(r.Run(*addr))
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	return router
}
