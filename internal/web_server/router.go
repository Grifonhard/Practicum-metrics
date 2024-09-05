package webserver

import (
	"context"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

func Middleware(next http.Handler, storage *storage.MemStorage)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		//проверяем запрос
		if r.Method != http.MethodPost{
			http.Error(w, "Just POST allow", http.StatusBadRequest)
		}
		//добавляем контекст в реквест
		ctx := context.WithValue(r.Context(), STORAGE_KEY, storage)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}