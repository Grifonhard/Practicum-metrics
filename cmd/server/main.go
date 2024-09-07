package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	web "github.com/Grifonhard/Practicum-metrics/internal/web_server"
)

func main() {
	mux := http.NewServeMux()
	stor := storage.New()

	mux.Handle("/update/{type}/{name}/{value}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			//добавляем контекст в реквест
			ctx := context.WithValue(r.Context(), web.STORAGE_KEY, stor)
			ctx = context.WithValue(ctx, web.TYPE_KEY, r.PathValue("type"))
			ctx = context.WithValue(ctx, web.NAME_KEY, r.PathValue("name"))
			ctx = context.WithValue(ctx, web.VALUE_KEY, r.PathValue("value"))

			web.Update(w, r.WithContext(ctx))	
		}))

	fmt.Printf("Server start localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}