package webclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"
	"log"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressBeforeSend(t *testing.T) {
	// Подготовка входных данных
	inputData := []byte(`{"id":"TestMetric","type":"gauge","value":123.456}`)
	inputBuffer := bytes.NewBuffer(inputData)

	// Вызов функции
	compressedBuffer, err := compressBeforeSend(inputBuffer)
	require.NoError(t, err, "compressBeforeSend вернула ошибку")

	// Распаковка данных для проверки
	reader, err := gzip.NewReader(compressedBuffer)
	require.NoError(t, err, "Не удалось создать gzip reader")
	defer reader.Close()

	decompressedData, err := io.ReadAll(reader)
	require.NoError(t, err, "Не удалось прочитать распакованные данные")

	// Проверка соответствия исходным данным
	assert.Equal(t, inputData, decompressedData, "Распакованные данные не соответствуют исходным")
}

func TestPrepareDataToSend(t *testing.T) {
	// Подготовка тестовых данных
	gaugeData := map[string]float64{
		"gaugeMetric1": 1.23,
		"gaugeMetric2": 4.56,
	}
	counterData := map[string]int64{
		"counterMetric1": 123,
		"counterMetric2": 456,
	}

	// Создаем канал для получения метрик
	ch := make(chan *Metrics, 4) // Буферизованный канал, чтобы избежать блокировки
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Вызов функции
	prepareDataToSend(gaugeData, counterData, ch, cancel)

	// Ожидаем завершения отправки метрик
	<-ctx.Done()
	close(ch)

	// Сбор метрик из канала
	var metrics []*Metrics
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Проверяем количество полученных метрик
	require.Len(t, metrics, 4, "Ожидается 4 метрики")

	// Ожидаемые метрики
	expectedMetrics := map[string]*Metrics{
		"gaugeMetric1": {
			ID:    "gaugeMetric1",
			MType: storage.TYPEGAUGE,
			Value: func() *float64 { v := 1.23; return &v }(),
		},
		"gaugeMetric2": {
			ID:    "gaugeMetric2",
			MType: storage.TYPEGAUGE,
			Value: func() *float64 { v := 4.56; return &v }(),
		},
		"counterMetric1": {
			ID:    "counterMetric1",
			MType: storage.TYPECOUNTER,
			Delta: func() *int64 { v := int64(123); return &v }(),
		},
		"counterMetric2": {
			ID:    "counterMetric2",
			MType: storage.TYPECOUNTER,
			Delta: func() *int64 { v := int64(456); return &v }(),
		},
	}

	// Создаем карту полученных метрик для удобства сравнения
	receivedMetricsMap := make(map[string]*Metrics)
	for _, metric := range metrics {
		receivedMetricsMap[metric.ID] = metric
	}

	// Сравниваем полученные метрики с ожидаемыми
	for id, expectedMetric := range expectedMetrics {
		actualMetric, exists := receivedMetricsMap[id]
		require.True(t, exists, "Метрика %s не была получена", id)
		assert.Equal(t, expectedMetric, actualMetric, "Метрика %s не соответствует ожидаемой", id)
	}
}

func TestSendMetric(t *testing.T) {
	// Создаем временный HTTP-сервер
	var receivedRequest *http.Request
	var receivedBody []byte

	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequest = r

		// Чтение сжатых данных из тела запроса
		compressedData, err := io.ReadAll(r.Body)
		require.NoError(t, err, "Не удалось прочитать тело запроса")

		// Распаковка данных
		gr, err := gzip.NewReader(bytes.NewReader(compressedData))
		require.NoError(t, err, "Не удалось создать gzip reader")
		defer gr.Close()

		receivedBody, err = io.ReadAll(gr)
		require.NoError(t, err, "Не удалось прочитать распакованные данные")

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Создаем реальный объект metgen.MetGen с тестовыми данными
	realMetGen := &metgen.MetGen{
		MetricsGauge: map[string]float64{
			"gaugeMetric": 123.456,
		},
		MetricsCounter: map[string]int64{
			"counterMetric": 789,
		},
	}

	// Вызов функции SendMetric с тестовым сервером и реальными метриками
	SendMetric(ts.URL, realMetGen, "", SENDSUBSEQUENCE)

	// Проверяем, что запрос был получен
	require.NotNil(t, receivedRequest, "Сервер не получил запрос")

	// Проверяем заголовки
	assert.Equal(t, "application/json", receivedRequest.Header.Get("Content-Type"))
	assert.Equal(t, "gzip", receivedRequest.Header.Get("Content-Encoding"))

	// Парсинг полученных метрик из JSON
	var metrics []Metrics
	decoder := json.NewDecoder(bytes.NewReader(receivedBody))
	for decoder.More() {
		var metric Metrics
		err := decoder.Decode(&metric)
		require.NoError(t, err, "Не удалось декодировать метрику")
		metrics = append(metrics, metric)
	}

	// Проверяем количество полученных метрик
	require.Len(t, metrics, 2, "Ожидается 2 метрики")

	// Ожидаемые метрики
	expectedMetrics := map[string]Metrics{
		"gaugeMetric": {
			ID:    "gaugeMetric",
			MType: storage.TYPEGAUGE,
			Value: func() *float64 { v := 123.456; return &v }(),
		},
		"counterMetric": {
			ID:    "counterMetric",
			MType: storage.TYPECOUNTER,
			Delta: func() *int64 { v := int64(789); return &v }(),
		},
	}

	// Сравниваем полученные метрики с ожидаемыми
	for _, metric := range metrics {
		expectedMetric, exists := expectedMetrics[metric.ID]
		require.True(t, exists, "Получена неожиданная метрика: %s", metric.ID)
		assert.Equal(t, expectedMetric, metric, "Метрика %s не соответствует ожидаемой", metric.ID)
	}
}