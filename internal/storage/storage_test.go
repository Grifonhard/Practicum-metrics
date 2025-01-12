package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage/fileio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockLogger struct{}

func (m *MockLogger) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type MockDB struct {
	mu             sync.Mutex
	metricsGauge   map[string]float64
	metricsCounter map[string][]float64
}

func NewMockDB() *MockDB {
	return &MockDB{
		metricsGauge:   make(map[string]float64),
		metricsCounter: make(map[string][]float64),
	}
}

func (m *MockDB) Begin() (*sql.Tx, error) {
	return nil, nil
}
func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}
func (m *MockDB) Close() error {
	return nil
}
func (m *MockDB) Conn(ctx context.Context) (*sql.Conn, error) {
	return nil, nil
}
func (m *MockDB) Driver() driver.Driver {
	return nil
}
func (m *MockDB) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *MockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *MockDB) Ping() error {
	return nil
}
func (m *MockDB) PingContext(ctx context.Context) error {
	return nil
}
func (m *MockDB) PingDB() error {
	return nil
}
func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	return nil, nil
}
func (m *MockDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}
func (m *MockDB) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockDB) QueryRow(query string, args ...any) *sql.Row {
	return nil
}
func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
func (m *MockDB) SetConnMaxIdleTime(d time.Duration) {}
func (m *MockDB) SetConnMaxLifetime(d time.Duration) {}
func (m *MockDB) SetMaxIdleConns(n int)              {}
func (m *MockDB) SetMaxOpenConns(n int)              {}
func (m *MockDB) Stats() sql.DBStats {
	return sql.DBStats{}
}

func (m *MockDB) CreateMetricsTable() error {
	return nil
}

func (m *MockDB) PushReplace(metricType, name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if metricType != TYPEGAUGE {
		return fmt.Errorf("PushReplace: unsupported metric type %s", metricType)
	}
	m.metricsGauge[name] = value
	return nil
}

func (m *MockDB) PushAdd(metricType, name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if metricType != TYPECOUNTER {
		return fmt.Errorf("PushAdd: unsupported metric type %s", metricType)
	}
	m.metricsCounter[name] = append(m.metricsCounter[name], value)
	return nil
}

func (m *MockDB) GetOneValue(metricType, name string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if metricType != TYPEGAUGE {
		return 0, fmt.Errorf("GetOneValue: unsupported metric type %s", metricType)
	}
	value, exists := m.metricsGauge[name]
	if !exists {
		return 0, ErrMetricNoData
	}
	return value, nil
}

func (m *MockDB) GetArrayValues(metricType, name string) ([]float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if metricType != TYPECOUNTER {
		return nil, fmt.Errorf("GetArrayValues: unsupported metric type %s", metricType)
	}
	values, exists := m.metricsCounter[name]
	if !exists {
		return nil, ErrMetricNoData
	}
	return values, nil
}

func (m *MockDB) List(metricOneValue string, metricArrayValues string) (map[string]float64, map[string][]float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	resultGauge := make(map[string]float64)
	resultCounter := make(map[string][]float64)

	metricTypes := []string{metricOneValue, metricArrayValues}

	for _, mt := range metricTypes {
		switch mt {
		case TYPEGAUGE:
			for k, v := range m.metricsGauge {
				resultGauge[k] = v
			}
		case TYPECOUNTER:
			for k, v := range m.metricsCounter {
				resultCounter[k] = v
			}
		default:
			return nil, nil, fmt.Errorf("List: unsupported metric type %s", mt)
		}
	}

	return resultGauge, resultCounter, nil
}

func TestNew(t *testing.T) {
	// Инициализируем моковый логгер.
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	t.Run("Создание файла в существующей директории без восстановления из бэкапа и без DB", func(t *testing.T) {
		tmpDir := t.TempDir()

		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		assert.NotNil(t, storage)
		defer storage.backupFile.Close()

		// Проверяем, что файл бэкапа создан.
		backupPath := filepath.Join(tmpDir, BACKUPFILENAME)
		_, statErr := os.Stat(backupPath)
		assert.NoError(t, statErr)

		// Проверяем, что ItemsGauge и ItemsCounter инициализированы пустыми.
		assert.NotNil(t, storage.ItemsGauge)
		assert.Empty(t, storage.ItemsGauge)
		assert.NotNil(t, storage.ItemsCounter)
		assert.Empty(t, storage.ItemsCounter)
	})

	t.Run("Создание файла в несуществующей директории с восстановлением из бэкапа", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "some", "subdir")

		// Создаём предварительный бэкап.
		backupFile, err := fileio.New(nestedDir, BACKUPFILENAME)
		require.NoError(t, err)
		data := &fileio.Data{
			ItemsGauge:   map[string]float64{"gauge1": 1.11},
			ItemsCounter: map[string][]float64{"counter1": {10, 20}},
		}
		err = backupFile.Write(data)
		require.NoError(t, err)
		backupFile.Close()

		// Теперь создаём MemStorage с восстановлением из бэкапа.
		storage, err := New(0, nestedDir, true, nil)
		require.NoError(t, err)
		assert.NotNil(t, storage)
		defer storage.backupFile.Close()

		// Проверяем, что данные восстановлены.
		assert.Equal(t, data.ItemsGauge, storage.ItemsGauge)
		assert.Equal(t, data.ItemsCounter, storage.ItemsCounter)
	})

	t.Run("Создание с использованием DB", func(t *testing.T) {
		tmpDir := t.TempDir()

		mockDB := NewMockDB()

		storage, err := New(0, tmpDir, false, mockDB)
		require.NoError(t, err)
		assert.NotNil(t, storage)
		defer storage.backupFile.Close()

		// Проверяем, что ItemsGauge и ItemsCounter не инициализированы.
		assert.Nil(t, storage.ItemsGauge)
		assert.Nil(t, storage.ItemsCounter)

		// Проверяем, что DB инициализирована.
		assert.Equal(t, mockDB, storage.DB)
	})

	t.Run("Ошибка при создании файла бэкапа", func(t *testing.T) {
		// Попытка создать файл в несуществующей директории без прав доступа.
		// На большинстве систем это может быть "/root" или "C:\Windows\System32".
		// Однако для кроссплатформенности лучше использовать путь, который не должен существовать.

		invalidPath := "/invalid_path_!@#$%^&*()"
		if os.PathSeparator == '\\' {
			invalidPath = "C:\\invalid_path_!@#$%^&*()"
		}

		storage, err := New(0, invalidPath, false, nil)
		assert.Error(t, err)
		assert.Nil(t, storage)
	})
}

func TestGet(t *testing.T) {
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	t.Run("Получение метрики из памяти", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Добавляем метрики вручную.
		storage.mu.Lock()
		storage.ItemsGauge["gauge1"] = 3.14
		storage.ItemsCounter["counter1"] = []float64{1, 2, 3}
		storage.mu.Unlock()

		// Получаем gauge.
		metric := &Metric{Type: TYPEGAUGE, Name: "gauge1"}
		value, err := storage.Get(metric)
		require.NoError(t, err)
		assert.Equal(t, 3.14, value)

		// Получаем counter.
		metricCounter := &Metric{Type: TYPECOUNTER, Name: "counter1"}
		counterValue, err := storage.Get(metricCounter)
		require.NoError(t, err)
		assert.Equal(t, 6.0, counterValue)
	})

	t.Run("Получение метрики из базы данных", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockDB := NewMockDB()

		// Инициализируем DB с некоторыми данными.
		mockDB.metricsGauge["gauge_db"] = 2.71
		mockDB.metricsCounter["counter_db"] = []float64{5, 15}

		storage, err := New(0, tmpDir, false, mockDB)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Получаем gauge из DB.
		metric := &Metric{Type: TYPEGAUGE, Name: "gauge_db"}
		value, err := storage.Get(metric)
		require.NoError(t, err)
		assert.Equal(t, 2.71, value)

		// Получаем counter из DB.
		metricCounter := &Metric{Type: TYPECOUNTER, Name: "counter_db"}
		counterValue, err := storage.Get(metricCounter)
		require.NoError(t, err)
		assert.Equal(t, 20.0, counterValue)
	})

	t.Run("Получение несуществующей метрики", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Попытка получить несуществующую метрику.
		metric := &Metric{Type: TYPEGAUGE, Name: "nonexistent"}
		value, err := storage.Get(metric)
		assert.Error(t, err)
		assert.Equal(t, 0.0, value)

		// Аналогично для counter.
		metricCounter := &Metric{Type: TYPECOUNTER, Name: "nonexistent_counter"}
		counterValue, err := storage.Get(metricCounter)
		assert.Error(t, err)
		assert.Equal(t, 0.0, counterValue)
	})

	t.Run("Ошибка при некорректном типе метрики", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Некорректный тип метрики.
		metric := &Metric{Type: "invalid_type", Name: "metric1"}
		value, err := storage.Get(metric)
		assert.Error(t, err)
		assert.Equal(t, 0.0, value)
	})

	t.Run("Ошибка при nil метрике", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		value, err := storage.Get(nil)
		assert.Error(t, err)
		assert.Equal(t, 0.0, value)
	})
}

func TestList(t *testing.T) {
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	t.Run("Список метрик из памяти", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Добавляем метрики вручную.
		storage.mu.Lock()
		storage.ItemsGauge["gauge1"] = 1.11
		storage.ItemsGauge["gauge2"] = 2.22
		storage.ItemsCounter["counter1"] = []float64{10, 20}
		storage.ItemsCounter["counter2"] = []float64{30, 40}
		storage.mu.Unlock()

		list, err := storage.List()
		require.NoError(t, err)

		expected := []string{
			"gauge1: 1.110000",
			"gauge2: 2.220000",
			"counter1: 10.000000, 20.000000",
			"counter2: 30.000000, 40.000000",
		}

		// Сортируем списки для сравнения без учёта порядка.
		assert.ElementsMatch(t, expected, list)
	})

	t.Run("Список метрик из базы данных", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockDB := NewMockDB()

		// Инициализируем DB с некоторыми данными.
		mockDB.metricsGauge["gauge_db1"] = 3.33
		mockDB.metricsGauge["gauge_db2"] = 4.44
		mockDB.metricsCounter["counter_db1"] = []float64{50, 60}
		mockDB.metricsCounter["counter_db2"] = []float64{70, 80}

		storage, err := New(0, tmpDir, false, mockDB)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		list, err := storage.List()
		require.NoError(t, err)

		expected := []string{
			"gauge_db1: 3.330000",
			"gauge_db2: 4.440000",
			"counter_db1: 50.000000, 60.000000",
			"counter_db2: 70.000000, 80.000000",
		}

		assert.ElementsMatch(t, expected, list)
	})

	t.Run("Список метрик при отсутствии метрик", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		list, err := storage.List()
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("Список метрик с некорректным типом метрики", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Добавляем некорректную метрику.
		storage.mu.Lock()
		storage.ItemsGauge["gauge1"] = 1.11
		// Предположим, что добавили метрику с некорректным типом.
		// Но в текущей реализации этого невозможно, так как структура разделяет типы.
		// Поэтому этот тест может быть пропущен или модифицирован.
		storage.mu.Unlock()

		list, err := storage.List()
		require.NoError(t, err)

		expected := []string{
			"gauge1: 1.110000",
		}

		assert.ElementsMatch(t, expected, list)
	})
}

func TestBackupLoop(t *testing.T) {
	assert.NoError(t, logger.Init(&MockLogger{}, 5))

	t.Run("Бэкап при получении сигнала из backupChan", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Добавляем метрики вручную.
		storage.mu.Lock()
		storage.ItemsGauge["gauge1"] = 5.55
		storage.ItemsCounter["counter1"] = []float64{100, 200}
		storage.mu.Unlock()

		// Запускаем BackupLoop в отдельной горутине.
		go storage.BackupLoop()

		// Отправляем сигнал в backupChan.
		storage.backupChan <- struct{}{}

		// Ждём немного, чтобы бэкап успел выполниться.
		time.Sleep(100 * time.Millisecond)

		// Читаем бэкап-файл и проверяем данные.
		gauges, counters, err := storage.backupFile.Read()
		require.NoError(t, err)
		assert.Equal(t, storage.ItemsGauge, gauges)
		assert.Equal(t, storage.ItemsCounter, counters)

		// Останавливаем BackupLoop через закрытие канала.
		storage.backupChan <- struct{}{}
		storage.backupFile.Close()
	})

	t.Run("Бэкап по тикеру", func(t *testing.T) {
		// Устанавливаем короткий интервал для теста.
		intervalBackup := 1 // секунда
		tmpDir := t.TempDir()
		storage, err := New(intervalBackup, tmpDir, false, nil)
		require.NoError(t, err)
		defer storage.backupFile.Close()

		// Добавляем метрики вручную.
		storage.mu.Lock()
		storage.ItemsGauge["gauge_tick"] = 6.66
		storage.ItemsCounter["counter_tick"] = []float64{300, 400}
		storage.mu.Unlock()

		// Запускаем BackupLoop в отдельной горутине.
		go storage.BackupLoop()

		// Ждём чуть больше интервала, чтобы бэкап успел выполниться.
		time.Sleep(1100 * time.Millisecond)

		// Читаем бэкап-файл и проверяем данные.
		gauges, counters, err := storage.backupFile.Read()
		require.NoError(t, err)
		assert.Equal(t, storage.ItemsGauge, gauges)
		assert.Equal(t, storage.ItemsCounter, counters)

		// Останавливаем BackupLoop через закрытие канала.
		storage.backupTicker.Stop()
		storage.backupFile.Close()
	})

	t.Run("Ошибка при записи бэкапа", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := New(0, tmpDir, false, nil)
		require.NoError(t, err)

		// Закрываем файл, чтобы запись завершилась с ошибкой.
		storage.backupFile.Close()

		// Запускаем BackupLoop в отдельной горутине.
		go storage.BackupLoop()

		// Отправляем сигнал в backupChan.
		storage.backupChan <- struct{}{}

		// Ждём немного, чтобы попытка записи произошла.
		time.Sleep(100 * time.Millisecond)

		// Останавливаем BackupLoop через закрытие канала.
		storage.backupChan <- struct{}{}
	})
}

func TestValidateAndConvert(t *testing.T) {
	t.Run("Успешная валидация и конвертация для gauge", func(t *testing.T) {
		method := http.MethodPost
		mType := TYPEGAUGE
		mName := "temperature"
		mValue := "23.5"

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		require.NoError(t, err)
		assert.Equal(t, mType, metric.Type)
		assert.Equal(t, mName, metric.Name)
		assert.Equal(t, 23.5, metric.Value)
	})

	t.Run("Успешная валидация и конвертация для counter", func(t *testing.T) {
		method := http.MethodPost
		mType := TYPECOUNTER
		mName := "requests"
		mValue := "100"

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		require.NoError(t, err)
		assert.Equal(t, mType, metric.Type)
		assert.Equal(t, mName, metric.Name)
		assert.Equal(t, 100.0, metric.Value)
	})

	t.Run("Метрика с методом GET", func(t *testing.T) {
		method := http.MethodGet
		mType := TYPEGAUGE
		mName := "humidity"
		mValue := "anyvalue" // должно быть проигнорировано

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		require.NoError(t, err)
		assert.Equal(t, 0.0, metric.Value)
	})

	t.Run("Ошибка при пустом типе метрики", func(t *testing.T) {
		method := http.MethodPost
		mType := ""
		mName := "metric1"
		mValue := "10"

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		assert.Error(t, err)
		assert.Nil(t, metric)
	})

	t.Run("Ошибка при пустом имени метрики", func(t *testing.T) {
		method := http.MethodPost
		mType := TYPEGAUGE
		mName := ""
		mValue := "10"

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		assert.Error(t, err)
		assert.Nil(t, metric)
	})

	t.Run("Ошибка при пустом значении метрики для POST", func(t *testing.T) {
		method := http.MethodPost
		mType := TYPECOUNTER
		mName := "metric2"
		mValue := ""

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		assert.Error(t, err)
		assert.Nil(t, metric)
	})

	t.Run("Метрика с некорректным типом", func(t *testing.T) {
		method := http.MethodPost
		mType := "invalid_type"
		mName := "metric3"
		mValue := "50"

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		assert.Error(t, err)
		assert.Nil(t, metric)
	})

	t.Run("Ошибка при некорректном значении метрики", func(t *testing.T) {
		method := http.MethodPost
		mType := TYPEGAUGE
		mName := "metric4"
		mValue := "not_a_float"

		metric, err := ValidateAndConvert(method, mType, mName, mValue)
		assert.Error(t, err)
		assert.Nil(t, metric)
	})
}

func TestPush(t *testing.T) {
	stor, err := New(0, "", false, nil)
	assert.NoError(t, err)

	metrics := []Metric{
		{
			Type:  "gauge",
			Name:  "name1",
			Value: 1.12,
		},
		{
			Type:  "gauge",
			Name:  "name1",
			Value: 2.24,
		},
		{
			Type:  "counter",
			Name:  "name2",
			Value: 3.36,
		},
		{
			Type:  "counter",
			Name:  "name2",
			Value: 4.48,
		},
	}

	go func() {
		for range stor.backupChan {
		}
	}()

	err = stor.Push(&metrics[0])
	assert.NoError(t, err)
	assert.Equal(t, stor.ItemsGauge[metrics[0].Name], metrics[0].Value)
	err = stor.Push(&metrics[1])
	assert.Equal(t, stor.ItemsGauge[metrics[0].Name], metrics[1].Value)
	assert.NoError(t, err)

	err = stor.Push(&metrics[2])
	assert.NoError(t, err)
	err = stor.Push(&metrics[3])
	assert.NoError(t, err)

	assert.Equal(t, stor.ItemsCounter[metrics[3].Name], []float64{metrics[2].Value, metrics[3].Value})
}

func TestMarshal(t *testing.T) {
	var item Metric
	item.Name = "test"
	item.Type = TYPEGAUGE
	item.Value = 1.11

	jn, err := json.Marshal(&item)
	assert.NoError(t, err)

	assert.Equal(t, `{"type":"gauge","id":"test","value":1.11}`, string(jn))
}

func TestUnmarshal(t *testing.T) {
	var item Metric
	err := json.Unmarshal([]byte(`{"id":"PollCount","type":"counter","delta":3}`), &item)
	assert.NoError(t, err)
	assert.Equal(t, `counter PollCount 3`, fmt.Sprintf("%s %s %0.f", item.Type, item.Name, item.Value))
}

func BenchmarkGetMemoryGauge(b *testing.B) {
	// Инициализируем моковый логгер.
	assert.NoError(b, logger.Init(&MockLogger{}, 5))

	tmpDir := b.TempDir()
	storage, err := New(0, tmpDir, false, nil)
	require.NoError(b, err)
	defer storage.backupFile.Close()

	// Добавляем метрику заранее.
	storage.mu.Lock()
	storage.ItemsGauge["benchmark_gauge"] = 1.23
	storage.mu.Unlock()

	metric := &Metric{
		Type: TYPEGAUGE,
		Name: "benchmark_gauge",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value, err := storage.Get(metric)
		require.NoError(b, err)
		assert.Equal(b, 1.23, value)
	}
}

func BenchmarkGetDatabaseGauge(b *testing.B) {
	// Инициализируем моковый логгер.
	assert.NoError(b, logger.Init(&MockLogger{}, 5))

	tmpDir := b.TempDir()
	mockDB := NewMockDB()
	mockDB.metricsGauge["benchmark_db_gauge"] = 1.23
	storage, err := New(0, tmpDir, false, mockDB)
	require.NoError(b, err)
	defer storage.backupFile.Close()

	metric := &Metric{
		Type: TYPEGAUGE,
		Name: "benchmark_db_gauge",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value, err := storage.Get(metric)
		require.NoError(b, err)
		assert.Equal(b, 1.23, value)
	}
}

func BenchmarkListMemory(b *testing.B) {
	// Инициализируем моковый логгер.
	assert.NoError(b, logger.Init(&MockLogger{}, 5))

	tmpDir := b.TempDir()
	storage, err := New(0, tmpDir, false, nil)
	require.NoError(b, err)
	defer storage.backupFile.Close()

	// Добавляем множество метрик заранее.
	storage.mu.Lock()
	for i := 0; i < 1000; i++ {
		storage.ItemsGauge[fmt.Sprintf("gauge_%d", i)] = float64(i)
		storage.ItemsCounter[fmt.Sprintf("counter_%d", i)] = []float64{float64(i), float64(i * 2)}
	}
	storage.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list, err := storage.List()
		require.NoError(b, err)
		assert.Len(b, list, 2000)
	}
}

func BenchmarkListDatabase(b *testing.B) {
	// Инициализируем моковый логгер.
	assert.NoError(b, logger.Init(&MockLogger{}, 5))

	tmpDir := b.TempDir()
	mockDB := NewMockDB()

	// Добавляем множество метрик в моковую DB.
	for i := 0; i < 1000; i++ {
		mockDB.metricsGauge[fmt.Sprintf("gauge_db_%d", i)] = float64(i)
		mockDB.metricsCounter[fmt.Sprintf("counter_db_%d", i)] = []float64{float64(i), float64(i * 2)}
	}

	storage, err := New(0, tmpDir, false, mockDB)
	require.NoError(b, err)
	defer storage.backupFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list, err := storage.List()
		require.NoError(b, err)
		assert.Len(b, list, 2000)
	}
}

func BenchmarkBackupLoopMemory(b *testing.B) {
	// Инициализируем моковый логгер.
	assert.NoError(b, logger.Init(&MockLogger{}, 5))

	tmpDir := b.TempDir()
	storage, err := New(0, tmpDir, false, nil)
	require.NoError(b, err)
	defer storage.backupFile.Close()

	// Добавляем метрики заранее.
	storage.mu.Lock()
	for i := 0; i < 1000; i++ {
		storage.ItemsGauge[fmt.Sprintf("gauge_%d", i)] = float64(i)
		storage.ItemsCounter[fmt.Sprintf("counter_%d", i)] = []float64{float64(i), float64(i * 2)}
	}
	storage.mu.Unlock()

	// Запускаем BackupLoop в отдельной горутине.
	go storage.BackupLoop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Отправляем сигнал для бэкапа.
		select {
		case storage.backupChan <- struct{}{}:
		default:
			// Если канал забит, пропускаем отправку.
		}
	}

	// Останавливаем BackupLoop через закрытие канала.
	storage.backupChan <- struct{}{}
	storage.backupFile.Close()
}

func BenchmarkBackupLoopDatabase(b *testing.B) {
	// Инициализируем моковый логгер.
	assert.NoError(b, logger.Init(&MockLogger{}, 5))

	tmpDir := b.TempDir()
	mockDB := NewMockDB()

	// Добавляем метрики в моковую DB.
	for i := 0; i < 1000; i++ {
		mockDB.metricsGauge[fmt.Sprintf("gauge_db_%d", i)] = float64(i)
		mockDB.metricsCounter[fmt.Sprintf("counter_db_%d", i)] = []float64{float64(i), float64(i * 2)}
	}

	storage, err := New(0, tmpDir, false, mockDB)
	require.NoError(b, err)
	defer storage.backupFile.Close()

	// Запускаем BackupLoop в отдельной горутине.
	go storage.BackupLoop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Отправляем сигнал для бэкапа.
		select {
		case storage.backupChan <- struct{}{}:
		default:
			// Если канал забит, пропускаем отправку.
		}
	}

	// Останавливаем BackupLoop через закрытие канала.
	storage.backupChan <- struct{}{}
	storage.backupFile.Close()
}
