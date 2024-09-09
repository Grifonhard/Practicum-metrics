package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/caarlos0/env/v10"
	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_ADDR = "localhost:8080"
)

type CFG struct{
	Addr string `env:"ADDRESS"`
}

func main() {
	addr := flag.String("a", DEFAULT_ADDR, "server port")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil{
		log.Fatal(err)
	}

	if cfg.Addr != ""{
		*addr = cfg.Addr
	}

	stor := storage.New()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.POST("/update/:type/:name/:value", web.Middleware(), web.Update(stor))
	r.GET("/value/:type/:name", web.Middleware(), web.Get(stor))
	r.GET("/", web.List(stor))

	fmt.Printf("Server start %s\n", *addr)
	log.Fatal(r.Run(*addr))
}