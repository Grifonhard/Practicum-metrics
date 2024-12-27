package metgen

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger реализует интерфейс логгера для тестов.
type MockLogger struct{}

func (m *MockLogger) Write(p []byte) (n int, err error) {
	// Для тестов можно просто игнорировать вывод или записывать в срез для проверки.
	return len(p), nil
}

func TestNew(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)
	assert.NotNil(t, mg.MetricsGauge)
	assert.NotNil(t, mg.MetricsCounter)
	assert.Equal(t, 0, len(mg.MetricsGauge))
	assert.Equal(t, 0, len(mg.MetricsCounter))
}

func TestRenewSuccess(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	// Выполняем Renew и ожидаем успешного завершения.
	err := mg.Renew()
	require.NoError(t, err)

	// Проверяем, что метрики обновлены.
	gg, cntr, err := mg.Collect()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(gg), 1)    // Должны быть собраны хотя бы несколько метрик.
	assert.GreaterOrEqual(t, len(cntr), 1) // Должны быть собраны хотя бы несколько метрик.
	assert.Equal(t, int64(1), cntr["PollCount"])
}

func TestRenewWithError(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	// Сохраняем оригинальную функцию и восстанавливаем её после теста.
	originalGetGopsutilMetricsFunc := getGopsutilMetricsFunc
	defer func() { getGopsutilMetricsFunc = originalGetGopsutilMetricsFunc }()

	// Переопределяем функцию, чтобы она возвращала ошибку и закрывала канал output.
	getGopsutilMetricsFunc = func(ctx context.Context, output chan OneMetric, errChan chan error) {
		errChan <- errors.New("mocked error in getGopsutilMetrics")
		close(output)
		close(errChan)
	}

	mg := New()
	require.NotNil(t, mg)

	// Выполняем Renew и ожидаем ошибки.
	err := mg.Renew()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mocked error in getGopsutilMetrics")
}

func TestCollect(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	// Добавляем некоторые метрики вручную.
	mg.mu.Lock()
	mg.MetricsGauge["test_gauge"] = 42.0
	mg.MetricsCounter["test_counter"] = 100
	mg.mu.Unlock()

	gg, cntr, err := mg.Collect()
	require.NoError(t, err)

	assert.Equal(t, 42.0, gg["test_gauge"])
	assert.Equal(t, int64(100), cntr["test_counter"])
}

func TestCollectGaugeToChan(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	// Добавляем некоторые метрики вручную.
	mg.mu.Lock()
	mg.MetricsGauge["test_gauge1"] = 10.0
	mg.MetricsGauge["test_gauge2"] = 20.0
	mg.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output := make(chan OneMetric)
	errChan := make(chan error)

	go mg.CollectGaugeToChan(ctx, output, errChan)

	collected := make(map[string]float64)
	for metric := range output {
		collected[metric.Name] = metric.Metric
	}

	// Проверяем, что все метрики были собраны.
	assert.Equal(t, 2, len(collected))
	assert.Equal(t, 10.0, collected["test_gauge1"])
	assert.Equal(t, 20.0, collected["test_gauge2"])
}

func TestCollectCounterToChan(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	// Добавляем некоторые метрики вручную.
	mg.mu.Lock()
	mg.MetricsCounter["test_counter1"] = 100
	mg.MetricsCounter["test_counter2"] = 200
	mg.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output := make(chan OneMetric)
	errChan := make(chan error)

	go mg.CollectCounterToChan(ctx, output, errChan)

	collected := make(map[string]float64)
	for metric := range output {
		collected[metric.Name] = metric.Metric
	}

	// Проверяем, что все метрики были собраны.
	assert.Equal(t, 2, len(collected))
	assert.Equal(t, float64(100), collected["test_counter1"])
	assert.Equal(t, float64(200), collected["test_counter2"])
}

func TestGetStandartMetrics(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	err := mg.Renew()
	require.NoError(t, err)

	gg, _, err := mg.Collect()
	require.NoError(t, err)

	// Проверяем наличие некоторых стандартных метрик.
	standardMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}

	for _, metric := range standardMetrics {
		_, exists := gg[metric]
		assert.True(t, exists, fmt.Sprintf("Metric %s should exist", metric))
	}
}

func TestGetGopsutilMetrics(t *testing.T) {
	// Аналогично TestGetStandartMetrics, тестирование происходит через Renew.
	// Проверяем наличие метрик, собранных gopsutil.

	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	err := mg.Renew()
	require.NoError(t, err)

	gg, _, err := mg.Collect()
	require.NoError(t, err)

	// Проверяем наличие метрик, собранных gopsutil.
	gopsutilMetrics := []string{
		"TotalMemory", "FreeMemory", "CpuUtilization",
	}

	for _, metric := range gopsutilMetrics {
		_, exists := gg[metric]
		assert.True(t, exists, fmt.Sprintf("Metric %s should exist", metric))
	}
}

func TestRenewConcurrency(t *testing.T) {
	// Этот тест проверяет, что Renew корректно обрабатывает параллельные обновления.

	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	mg := New()
	require.NotNil(t, mg)

	// Запускаем несколько горутин, которые одновременно вызывают Renew.
	var wg sync.WaitGroup
	numRoutines := 5
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func() {
			defer wg.Done()
			err := mg.Renew()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	// Проверяем, что метрики обновлены корректно.
	_, cntr, err := mg.Collect()
	require.NoError(t, err)

	// Проверяем, что PollCount увеличился на количество вызовов Renew.
	expectedPollCount := int64(numRoutines)
	assert.Equal(t, expectedPollCount, cntr["PollCount"])
}

func TestRenewContextCancellation(t *testing.T) {
	// Инициализируем моковый логгер.
    assert.NoError(t, logger.Init(&MockLogger{}, 5))

    // Сохраняем оригинальную функцию и восстанавливаем её после теста.
    originalGetGopsutilMetricsFunc := getGopsutilMetricsFunc
    defer func() { getGopsutilMetricsFunc = originalGetGopsutilMetricsFunc }()

    // Переопределяем функцию, чтобы она возвращала ошибку и закрывала канал output.
    getGopsutilMetricsFunc = func(ctx context.Context, output chan OneMetric, errChan chan error) {
        errChan <- errors.New("mocked error in getGopsutilMetrics")
        close(output)
		close(errChan)
    }

    mg := New()
    require.NotNil(t, mg)

    // Выполняем Renew и ожидаем ошибки.
    err := mg.Renew()
    require.Error(t, err)
    assert.Contains(t, err.Error(), "mocked error in getGopsutilMetrics")

    // Проверяем, что PollCount не увеличился.
    _, cntr, err := mg.Collect()
    require.NoError(t, err)
    assert.Equal(t, int64(0), cntr["PollCount"])
}
