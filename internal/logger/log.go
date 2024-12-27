package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// Init инициализация глобального логгера 
// 	level from 0 to 5
// 	level 0 - нет логирования
// 	level 5 - Debug level
func Init(output io.Writer, level int) error {
	log := logrus.New()

	if output != nil {
		log.SetOutput(output)
	} else {
		log.SetOutput(os.Stdout)
	}

	log.SetLevel(logrus.Level(level))

	logger = log

	return nil
}

// Info логгирование уровня info
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Error логгирования уровня Error
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Warn логгирование уровня Warn
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Debug логгирование уровня Debug
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// WithFields добавление полей в логи
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}
