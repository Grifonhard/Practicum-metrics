package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

//level from 0 to 5
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
