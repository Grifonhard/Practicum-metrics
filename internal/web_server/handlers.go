package webserver

import (
	"fmt"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

const (
	PARAMS_AMOUNT = 3
	STORAGE_KEY   = "storage"
	TYPE_KEY = "type"
	NAME_KEY = "name"
	VALUE_KEY = "value"
)

func Update(w http.ResponseWriter, r *http.Request) {
	//извлекаем данные из контекста
	stor, ok := r.Context().Value(STORAGE_KEY).(*storage.MemStorage)
	if !ok {
		http.Error(w, "Storage not found in context", http.StatusInternalServerError)
		return
	}
	mType, ok := r.Context().Value(TYPE_KEY).(string)
	if !ok {
		http.Error(w, "Type not found in context", http.StatusInternalServerError)
		return
	}
	mName, ok := r.Context().Value(NAME_KEY).(string)
	if !ok {
		http.Error(w, "Name not found in context", http.StatusInternalServerError)
		return
	}
	mValue, ok := r.Context().Value(VALUE_KEY).(string)
	if !ok {
		http.Error(w, "Value not found in context", http.StatusInternalServerError)
		return
	}

	//сохраняем данные
	err := stor.Push(mName, mValue, mType)
	if err != nil {
		http.Error(w, "Fail while push", http.StatusInternalServerError)
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
