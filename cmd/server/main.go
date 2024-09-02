package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/Grifonhard/Practicum-metrics/internal/web"
)

func main() {
	mux := http.NewServeMux()
	storage := storage.New()

	mux.Handle("/update/", web.Middleware(http.HandlerFunc(web.Update), storage))

	fmt.Printf("Server start localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
