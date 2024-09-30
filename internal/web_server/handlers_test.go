package webserver

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	//подготовка
	stor := storage.New()

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
	stor := storage.New()
	stor.ItemsCounter = map[string][]float64{
		"testcounter": {1.11, 2.22},
	}
	stor.ItemsGauge = map[string]float64{
		"testgauge": 3.33,
	}

	err := logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	getJSONPath := "/value/"

	router := gin.Default()

	router.POST("/value/", ReqRespLogger(), GetJSON(stor))

	buf := bytes.NewBuffer([]byte(`{"id":"testgauge","type":"gauge"}`))

	r := httptest.NewRequest(http.MethodPost, getJSONPath, buf)
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	fmt.Println(w.Body.String())
}

func TestDataExtraction(t *testing.T) {
	// Устанавливаем режим тестирования для Gin
	gin.SetMode(gin.TestMode)
	
	// Создаем новый экземпляр роутера
	r := gin.New()
	
	// Регистрируем middleware DataExtraction
	r.Use(DataExtraction())
	
	// Создаем тестовый обработчик для POST /update
	r.POST("/update", func(c *gin.Context) {
		metricType, exists := c.Get(METRICTYPE)
		if !exists {
			c.String(http.StatusInternalServerError, "METRICTYPE not set")
			return
		}
		
		// В зависимости от типа метрики возвращаем разный контент
		if metricType == METRICTYPEJSON {
			response := gin.H{"metric_type": metricType, "status": "success"}
			c.JSON(http.StatusOK, response)
		} else {
			c.String(http.StatusOK, "Metric type: %s", metricType)
		}
	})
	
	// Вспомогательная функция для проверки, является ли ответ gzip-сжатым
	isGzipped := func(resp *httptest.ResponseRecorder) bool {
		encoding := resp.Header().Get("Content-Encoding")
		return strings.Contains(encoding, "gzip")
	}
	
	// Вспомогательная функция для декомпрессии gzip-ответа
	decompressGzip := func(data []byte) (string, error) {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return "", err
		}
		defer reader.Close()
		
		var decompressed bytes.Buffer
		_, err = io.Copy(&decompressed, reader)
		if err != nil {
			return "", err
		}
		return decompressed.String(), nil
	}
	
	// 1. POST-запрос без сжатия и без Accept-Encoding: gzip
	t.Run("POST without compression and without Accept-Encoding", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`{"key": "value"}`)
		req, err := http.NewRequest(http.MethodPost, "/update", reqBody)
		assert.NoError(t, err)
		
		// Устанавливаем заголовок Content-Type
		req.Header.Set("Content-Type", "application/json")
		
		// Выполняем запрос
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// Проверяем статус-код
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Проверяем, что ответ не сжат
		assert.False(t, isGzipped(w), "Response should not be gzipped")
		
		// Проверяем тело ответа
		expectedBody := "Metric type: default"
		assert.Equal(t, expectedBody, w.Body.String())
	})
	
	// 2. POST-запрос с сжатием данных и с запросом сжатого ответа
	t.Run("POST with gzip compression and Accept-Encoding: gzip", func(t *testing.T) {
		// Сжимаем тело запроса
		var gzippedBody bytes.Buffer
		gz := gzip.NewWriter(&gzippedBody)
		_, err := gz.Write([]byte(`{"key": "value"}`))
		assert.NoError(t, err)
		gz.Close()
		
		req, err := http.NewRequest(http.MethodPost, "/update", &gzippedBody)
		assert.NoError(t, err)
		
		// Устанавливаем заголовки Content-Encoding и Content-Type
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		
		// Устанавливаем заголовок Accept-Encoding
		req.Header.Set("Accept-Encoding", "gzip")
		
		// Выполняем запрос
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// Проверяем статус-код
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Проверяем, что ответ сжат
		assert.True(t, isGzipped(w), "Response should be gzipped")
		
		// Декомпрессируем ответ
		decompressedBody, err := decompressGzip(w.Body.Bytes())
		assert.NoError(t, err)
		
		// Проверяем тело ответа
		var response map[string]string
		err = json.Unmarshal([]byte(decompressedBody), &response)
		assert.NoError(t, err)
		
		expectedResponse := map[string]string{
			"metric_type": "json",
			"status":      "success",
		}
		assert.Equal(t, expectedResponse, response)
	})
	
	// 3. POST-запрос с сжатием данных без запроса сжатого ответа
	t.Run("POST with gzip compression without Accept-Encoding: gzip", func(t *testing.T) {
		// Сжимаем тело запроса
		var gzippedBody bytes.Buffer
		gz := gzip.NewWriter(&gzippedBody)
		_, err := gz.Write([]byte(`{"key": "value"}`))
		assert.NoError(t, err)
		gz.Close()
		
		req, err := http.NewRequest(http.MethodPost, "/update", &gzippedBody)
		assert.NoError(t, err)
		
		// Устанавливаем заголовок Content-Encoding и Content-Type
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		
		// Не устанавливаем заголовок Accept-Encoding
		
		// Выполняем запрос
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// Проверяем статус-код
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Проверяем, что ответ не сжат
		assert.False(t, isGzipped(w), "Response should not be gzipped")
		
		// Проверяем тело ответа
		expectedBody := "Metric type: json"
		assert.Equal(t, expectedBody, w.Body.String())
	})
	
	// 4. POST-запрос без сжатия данных, но с запросом сжатого ответа
	t.Run("POST without compression but with Accept-Encoding: gzip", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`{"key": "value"}`)
		req, err := http.NewRequest(http.MethodPost, "/update", reqBody)
		assert.NoError(t, err)
		
		// Устанавливаем заголовки Content-Type и Accept-Encoding
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		
		// Выполняем запрос
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// Проверяем статус-код
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Проверяем, что ответ сжат
		assert.True(t, isGzipped(w), "Response should be gzipped")
		
		// Декомпрессируем ответ
		decompressedBody, err := decompressGzip(w.Body.Bytes())
		assert.NoError(t, err)
		
		// Проверяем тело ответа
		var response map[string]string
		err = json.Unmarshal([]byte(decompressedBody), &response)
		assert.NoError(t, err)
		
		expectedResponse := map[string]string{
			"metric_type": "json",
			"status":      "success",
		}
		assert.Equal(t, expectedResponse, response)
	})
	
	// 5. POST-запрос с некорректным сжатием данных
	t.Run("POST with invalid gzip compression", func(t *testing.T) {
		// Некорректные сжатые данные (обычная строка)
		reqBody := bytes.NewBufferString(`{"key": "value"}`)
		req, err := http.NewRequest(http.MethodPost, "/update", reqBody)
		assert.NoError(t, err)
		
		// Устанавливаем заголовок Content-Encoding: gzip, но данные не сжаты
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		
		// Выполняем запрос
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// Проверяем статус-код
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		// Проверяем содержимое тела ответа
		assert.Contains(t, w.Body.String(), "fail while create decompress request")
		
		// Проверяем заголовки
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	})
}