package webserver

import (
	"net/http"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type respInfo struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	gin.ResponseWriter
	respInfo *respInfo
}

func (lw *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(data)
	lw.respInfo.size = size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.respInfo.status = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

func DataExtraction() gin.HandlerFunc {
	return func(c *gin.Context) {
		item, err := storage.ValidateAndConvert(c.Request.Method, c.Param("type"), c.Param("name"), c.Param("value"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.Set(METRICKEY, item)

		c.Next()
	}
}

func ReqRespLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		respInfo := &respInfo{}

		lw := &loggingResponseWriter{
			ResponseWriter: c.Writer,
			respInfo:       respInfo,
		}

		c.Writer = lw
		c.Next()

		duration := time.Since(start)

		logger.WithFields(logrus.Fields{
			"URL":          c.Request.URL,
			"method":       c.Request.Method,
			"lead time ms": duration.Milliseconds(),
			"status":       lw.respInfo.status,
			"size":         lw.respInfo.size,
		}).Info()
	}
}
