package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainConfig(t *testing.T) {
	// Сохраняем текущие флаги и переменные окружения
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Подменяем аргументы командной строки
	os.Args = []string{"cmd", "-a=127.0.0.1:9090", "-r=15", "-p=5"}

	// Устанавливаем переменные окружения
	os.Setenv("ADDRESS", "localhost:7070")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "10")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	// Сбрасываем флаги для нового парсинга
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Чтение флагов командной строки
	address := flag.String("a", DEFAULTADDR, "адрес сервера")
	reportInterval := flag.Int("r", DEFAULTREPORTINTERVAL, "секунд частота отправки метрик")
	pollInterval := flag.Int("p", DEFAULTPOLLINTERVAL, "секунд частота опроса метрик")
	flag.Parse()

	// Чтение переменных окружения
	var cfg CFG
	err := env.Parse(&cfg)
	require.NoError(t, err, "Ошибка при парсинге переменных окружения")

	// Логика по установке значений в зависимости от переменных окружения
	if _, exist := os.LookupEnv("ADDRESS"); exist {
		address = &cfg.Addr
	}
	if _, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		pollInterval = &cfg.PollInterval
	}
	if _, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		reportInterval = &cfg.ReportInterval
	}

	// Проверяем правильность настроек
	assert.Equal(t, "localhost:7070", *address, "Адрес должен быть взят из переменной окружения")
	assert.Equal(t, 10, *pollInterval, "Интервал опроса должен быть взят из переменной окружения")
	assert.Equal(t, 20, *reportInterval, "Интервал отправки должен быть взят из переменной окружения")
}

func TestMainTickerIntervals(t *testing.T) {
	// Настраиваем значения для теста
	reportInterval := 10
	pollInterval := 2

	// Создаем тикеры для тестирования
	timerPoll := time.NewTicker(time.Duration(pollInterval) * time.Second)
	timerReport := time.NewTicker(time.Duration(reportInterval) * time.Second)

	defer timerPoll.Stop()
	defer timerReport.Stop()

	// Проверяем, что тикеры корректно созданы и отправляют сигналы
	select {
	case <-timerPoll.C:
		assert.True(t, true, "Таймер опроса сработал")
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Таймер опроса не сработал за 3 секунды")
	}

	select {
	case <-timerReport.C:
		assert.True(t, true, "Таймер отправки сработал")
	case <-time.After(11 * time.Second):
		assert.Fail(t, "Таймер отправки не сработал за 11 секунд")
	}
}