package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func Init() error {
	log := logrus.New()

	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(logrus.InfoLevel)

	logger = log

	return nil
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}
