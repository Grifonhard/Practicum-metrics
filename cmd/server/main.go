package main

import (
	"fmt"
	"log"
	"flag"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/gin-gonic/gin"
)

func main() {
	port := flag.String("a", "8080", "server port")

	flag.Parse()

	stor := storage.New()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.POST("/update/:type/:name/:value", web.Middleware(), web.Update(stor))
	r.GET("/value/:type/:name", web.Middleware(), web.Get(stor))
	r.GET("/", web.List(stor))

	fmt.Printf("Server start localhost:%s\n", *port)
	log.Fatal(r.Run(":" + *port))
}