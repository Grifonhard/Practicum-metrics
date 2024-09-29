package webserver

import (
	"compress/gzip"
	"net/http"
	"strings"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
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

type decompressRequest struct {
	*http.Request
	gzipReader *gzip.Reader
}

func NewDecompressRequest(req *http.Request) (*decompressRequest, error) {
	gzipReader, err := gzip.NewReader(req.Body)
	if err != nil {
		return nil, err
	}
	return &decompressRequest{
		Request:    req,
		gzipReader: gzipReader,
	}, nil
}

func (d *decompressRequest) Read(p []byte) (n int, err error) {
	return d.gzipReader.Read(p)
}

func (d *decompressRequest) Close() error {
	if err := d.gzipReader.Close(); err != nil {
		return err
	}
	return d.Request.Body.Close()
}

type compressResponseWriter struct {
}

func DataExtraction() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && strings.Contains(c.Request.URL.Path, "/update") && strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
			c.Set(METRICTYPE, METRICTYPEJSON)

			c.Next()
		} else {

			c.Set(METRICTYPE, METRICTYPEDEFAULT)

			c.Next()
		}
	}
}
