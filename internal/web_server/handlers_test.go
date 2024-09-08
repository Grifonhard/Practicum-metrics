package webserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T){
	//подготовка
	stor := storage.New()

	updatePath := "/update"
	
	name := "name"
	value := 1.14
	itemType1 := storage.TYPE1
	itemType2 := storage.TYPE2

	item1 := storage.Metric{
		Type: itemType1,
		Name: name,
		Value: value,
	}

	item2 := storage.Metric{
		Type: itemType2,
		Name: name,
		Value: value,
	}

	urlSuccess1 :=  fmt.Sprintf("%s/%s/%s/%.2f", updatePath, item1.Type, item1.Name, item1.Value)
	urlSuccess2 :=  fmt.Sprintf("%s/%s/%s/%.2f", updatePath, item2.Type, item2.Name, item2.Value)
	urlWrongType := fmt.Sprintf("%s/%s/%s/%.2f", updatePath, "wrong", item1.Name, item1.Value)
	urlWrongValue := fmt.Sprintf("%s/%s/%s/%s", updatePath, item1.Type, item1.Name, "wrong")

	methodSuccess := http.MethodPost
	methodWrong := http.MethodGet

	router := gin.Default()

	router.POST("/update/:type/:name/:value", Middleware(), Update(stor))

	tests := []struct{
		url string
		method string
		waitErr bool
		message string
	}{
		{urlSuccess1, methodSuccess, false, "type 1 success"},
		{urlSuccess2, methodSuccess, false, "type 2 success"},
		{urlWrongType, methodSuccess, true, "wrong type"},
		{urlWrongValue, methodSuccess, true, "wrong value"},
		{urlSuccess1, methodWrong, true, "wrong method"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.url, nil)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			if tt.waitErr {
				assert.NotEqual(t, http.StatusOK, w.Code, tt.message)
			} else {
				assert.Equal(t, http.StatusOK, w.Code, tt.message)
			}
		})
	}
}