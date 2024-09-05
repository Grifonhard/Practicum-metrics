package webserver
import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

const(
	PARAMS_AMOUNT = 3
	STORAGE_KEY = "storage"
)

func Update(w http.ResponseWriter, r *http.Request){	
	//извлекаем параметры из запроса
	paramsString := strings.TrimPrefix(r.URL.Path, "/update/")
	params := strings.Split(paramsString, "/")
	if len(params) != PARAMS_AMOUNT{
		http.Error(w, fmt.Sprintf("number of request parameters does not match expected, "+ 
		"expected: %d, received: %d", PARAMS_AMOUNT, len(params)), http.StatusBadRequest)
		return
	} else if params[1] == ""{
		http.Error(w, "Metric name not found in request", http.StatusNotFound)
		return
	}
	//извлекаем хранилку из контекста
	storage, ok := r.Context().Value(STORAGE_KEY).(*storage.MemStorage)
	if !ok{
		http.Error(w, "No storage found in context", http.StatusInternalServerError)
		return
	}
	//сохраняем данные
	mName := params[1]
	mValue := params[2]
	mType := params[0]
	err := storage.Push(mName, mValue, mType)
	if err != nil{
		http.Error(w, "Fail while push", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Сontent-Length", fmt.Sprint(len("Success")))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Success"))
	if err !=nil{
		http.Error(w, "Fail while write", http.StatusInternalServerError)
	}
}