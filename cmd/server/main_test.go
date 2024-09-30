package main

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/caarlos0/env/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	// Сохраняем текущие аргументы и флаги
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Подменяем аргументы командной строки
	os.Args = []string{"cmd", "-a=127.0.0.1:9090", "-i=100", "-f=/tmp/test", "-r=false"}

	// Устанавливаем переменные окружения
	os.Setenv("ADDRESS", "localhost:7070")
	os.Setenv("STORE_INTERVAL", "200")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/envtest")
	os.Setenv("RESTORE", "true") // Устанавливаем переменную окружения в true
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
	}()

	// Сбросим флаги для нового парсинга
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Запускаем main логику для теста
	addr := flag.String("a", DEFAULTADDR, "server port")
	storeInterval := flag.Int("i", DEFAULTSTOREINTERVAL, "backup interval")
	fileStoragePath := flag.String("f", "", "file storage path")
	restore := flag.Bool("r", DEFAULTRESTORE, "restore from backup")
	flag.Parse()

	// Парсинг переменных окружения
	var cfg CFG
	err := env.Parse(&cfg)
	require.NoError(t, err, "Ошибка при парсинге переменных окружения")

	// Применяем логику из main
	if _, exist := os.LookupEnv("ADDRESS"); exist {
		addr = &cfg.Addr
	}
	if _, exist := os.LookupEnv("STORE_INTERVAL"); exist  {
		storeInterval = &cfg.StoreInterval
	}
	if _, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist  {
		fileStoragePath = &cfg.FileStoragePath
	}
	if _, exist := os.LookupEnv("RESTORE"); exist  {
		restore = &cfg.Restore
	}

	// Проверяем правильность настроек
	assert.Equal(t, "localhost:7070", *addr, "Адрес должен быть взят из переменной окружения")
	assert.Equal(t, 200, *storeInterval, "Интервал должен быть взят из переменной окружения")
	assert.Equal(t, "/tmp/envtest", *fileStoragePath, "Путь должен быть взят из переменной окружения")
	assert.True(t, *restore, "Восстановление должно быть взято из переменной окружения")
}

func TestLoggerInit(t *testing.T) {
	var buf bytes.Buffer

	// Инициализация логгера
	err := logger.Init(&buf, 4)
	assert.NoError(t, err, "Логгер должен инициализироваться без ошибок")

	// Проверка записи логгера
	logger.Info("Test info message")
	assert.Contains(t, buf.String(), "Test info message", "Лог должен содержать сообщение 'Test info message'")
}

func TestStorageInit(t *testing.T) {
	// Инициализация хранилища
	storeInterval := 100
	fileStoragePath := "/tmp/teststorage"
	restore := true

	stor, err := storage.New(storeInterval, fileStoragePath, restore)
	require.NoError(t, err, "Хранилище должно инициализироваться без ошибок")

	assert.NotNil(t, stor, "Хранилище не должно быть nil")
}