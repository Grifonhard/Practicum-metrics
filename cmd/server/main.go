package main

import (
	"fmt"
	"log"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/gin-gonic/gin"
)

func main() {
	stor := storage.New()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.POST("/update/:type/:name/:value", web.Middleware(), web.Update(stor))
	r.GET("/value/:type/:name", web.Middleware(), web.Get(stor))
	r.GET("/", web.List(stor))

	fmt.Printf("Server start localhost:8080")
	log.Fatal(r.Run(":8012"))
}