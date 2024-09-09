package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/caarlos0/env/v10"
	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_PORT = "8080"
)

type CFG struct{
	Addr string `env:"ADDRESS"`
}

func main() {
	port := flag.String("a", DEFAULT_PORT, "server port")

	flag.Parse()

	var cfg CFG
	err := env.Parse(&cfg)
	if err != nil{
		log.Fatal(err)
	}

	if cfg.Addr != ""{
		*port, _ = strings.CutPrefix(cfg.Addr, "localhost:")
	}

	stor := storage.New()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.POST("/update/:type/:name/:value", web.Middleware(), web.Update(stor))
	r.GET("/value/:type/:name", web.Middleware(), web.Get(stor))
	r.GET("/", web.List(stor))

	fmt.Printf("Server start localhost:%s\n", *port)
	log.Fatal(r.Run(":" + *port))
}