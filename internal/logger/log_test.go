package logger

import (
	"bytes"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	err := Init(os.Stdout, 0)
	require.NoError(t, err, "Init должен проходить без ошибок")
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer

	// Инициализация логгера с выводом в буфер
	Init(&buf,4)

	// Вызываем функцию Info
	logger.Info("Test info message")

	// Проверяем, что сообщение попало в лог
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test info message", "Сообщение должно содержать 'Test info message'")
	assert.Contains(t, logOutput, "info", "Лог должен быть уровня 'info'")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer

	// Инициализация логгера с выводом в буфер
	Init(&buf,2)

	// Вызываем функцию Error
	logger.Error("Test error message")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test error message", "Сообщение должно содержать 'Test error message'")
	assert.Contains(t, logOutput, "error", "Лог должен быть уровня 'error'")
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer

	// Инициализация логгера с выводом в буфер
	Init(&buf,3)

	// Вызываем функцию Warn
	logger.Warn("Test warning message")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test warning message", "Сообщение должно содержать 'Test warning message'")
	assert.Contains(t, logOutput, "warning", "Лог должен быть уровня 'warning'")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer

	// Инициализация логгера с выводом в буфер
	Init(&buf,5)

	// Устанавливаем уровень Debug
	logger.Debug("Test debug message")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test debug message", "Сообщение должно содержать 'Test debug message'")
	assert.Contains(t, logOutput, "debug", "Лог должен быть уровня 'debug'")
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer

	// Инициализация логгера с выводом в буфер
	Init(&buf,4)

	// Используем WithFields для добавления полей
	logger.WithFields(logrus.Fields{
		"key": "value",
	}).Info("Test message with fields")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "key=value", "Лог должен содержать переданные поля")
	assert.Contains(t, logOutput, "Test message with fields", "Сообщение должно быть в логе")
	assert.Contains(t, logOutput, "info", "Лог должен быть уровня 'info'")
}
