package webserver

import (
	"fmt"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

const (
	PARAMS_AMOUNT = 3
	STORAGE_KEY   = "storage"
	METRIC_KEY = "metric"
)

func Update(w http.ResponseWriter, r *http.Request) {
	//проверяем запрос
	if r.Method != http.MethodPost {
		http.Error(w, "Just POST allow", http.StatusBadRequest)
		return
	}
	//извлекаем данные из контекста
	stor, ok := r.Context().Value(STORAGE_KEY).(*storage.MemStorage)
	if !ok {
		http.Error(w, "Storage not found in context", http.StatusInternalServerError)
		return
	}
	item, ok := r.Context().Value(METRIC_KEY).(*storage.Metric)
	if !ok {
		http.Error(w, "Metric not found in context", http.StatusInternalServerError)
		return
	}

	//сохраняем данные
	err := stor.Push(item)
	if err != nil {
		http.Error(w, fmt.Sprintf("Fail while push error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Сontent-Length", fmt.Sprint(len("Success")))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Success"))
	if err != nil {
		http.Error(w, "Fail while write", http.StatusInternalServerError)
	}
}