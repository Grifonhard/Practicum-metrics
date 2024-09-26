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
	DEFAULTADDR = "localhost:8080"
)

type CFG struct {
	Addr string `env:"ADDRESS"`
}

func main() {
	addr := flag.String("a", DEFAULTADDR, "server port")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Addr != "" {
		*addr = cfg.Addr
	}

	stor := storage.New()

	err = logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	r := initRouter()

	r.POST("/update", web.ReqRespLogger(), web.DataExtraction(), web.Update(stor))
	r.POST("/update/:type/:name/:value", web.ReqRespLogger(), web.DataExtraction(), web.Update(stor))
	r.GET("/value/:type/:name", web.ReqRespLogger(), web.DataExtraction(), web.Get(stor))
	r.GET("/", web.ReqRespLogger(), web.List(stor))

	fmt.Printf("Server start %s\n", *addr)
	log.Fatal(r.Run(*addr))
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	return router
}
