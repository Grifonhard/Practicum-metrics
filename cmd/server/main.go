package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
	"github.com/gin-gonic/gin"
)

func main() {
	stor := storage.New()

	r := gin.Default()

	r.POST("/update/:type/:name/:value", web.Middleware(), web.Update(stor))

	fmt.Printf("Server start localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}