package webserver

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	//подготовка
	stor, err := storage.New(0, "", false, nil)
	assert.NoError(t, err)

	go stor.BackupLoop()

	updatePath := "/update"

	name := "name"
	value := 1.14
	itemType1 := storage.TYPEGAUGE
	itemType2 := storage.TYPECOUNTER

	item1 := storage.Metric{
		Type:  itemType1,
		Name:  name,
		Value: value,
	}

	item2 := storage.Metric{
		Type:  itemType2,
		Name:  name,
		Value: value,
	}

	urlSuccess1 := fmt.Sprintf("%s/%s/%s/%.2f", updatePath, item1.Type, item1.Name, item1.Value)
	urlSuccess2 := fmt.Sprintf("%s/%s/%s/%.2f", updatePath, item2.Type, item2.Name, item2.Value)
	urlWrongType := fmt.Sprintf("%s/%s/%s/%.2f", updatePath, "wrong", item1.Name, item1.Value)
	urlWrongValue := fmt.Sprintf("%s/%s/%s/%s", updatePath, item1.Type, item1.Name, "wrong")

	methodSuccess := http.MethodPost
	methodWrong := http.MethodGet

	router := gin.Default()

	router.POST("/update/:type/:name/:value", DataExtraction(), Update(stor))

	tests := []struct {
		url     string
		method  string
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

func TestGetJSON(t *testing.T) {
	//подготовка
	stor, err := storage.New(0, "", false, nil)
	assert.NoError(t, err)

	go stor.BackupLoop()

	stor.ItemsCounter = map[string][]float64{
		"testcounter": {1.11, 2.22},
	}
	stor.ItemsGauge = map[string]float64{
		"testgauge": 3.33,
	}

	err = logger.Init(os.Stdout,4)
	if err != nil {
		log.Fatal(err)
	}

	getJSONPath := "/value/"

	router := gin.Default()

	router.POST("/value/", ReqRespLogger(""), GetJSON(stor))

	buf := bytes.NewBuffer([]byte(`{"id":"testgauge","type":"gauge"}`))

	r := httptest.NewRequest(http.MethodPost, getJSONPath, buf)
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	fmt.Println(w.Body.String())
}
