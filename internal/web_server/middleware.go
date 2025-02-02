package webserver

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// respInfo хранит информацию для логирования обращений к серверу
type respInfo struct {
	status int
	size   int
	hash   string // на случай если потребуется работа с псевдоаутентификацией
}

// loggingResponseWriter используется для логирования обращений к серверу
// обёртка для ResponseWriter
type loggingResponseWriter struct {
	gin.ResponseWriter
	respInfo *respInfo
	key      string // на случай если потребуется работа с псевдоаутентификацией
}

// Write записывает данные в ResponseWriter, считает хэш
func (lw *loggingResponseWriter) Write(data []byte) (int, error) {
	if lw.key != "" {
		lw.respInfo.hash = computeHMAC(data, lw.key)
	}
	size, err := lw.ResponseWriter.Write(data)
	lw.respInfo.size = size
	return size, err
}

// WriteHeader записывает заголовки в ResponseWrite, отправляет хэш в заголовке
func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.Header().Set("HashSHA256", lw.respInfo.hash)
	lw.respInfo.status = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

// ReqRespLogger логгирует реквесты и респонсы
func ReqRespLogger(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		respInfo := &respInfo{}

		lw := &loggingResponseWriter{
			ResponseWriter: c.Writer,
			respInfo:       respInfo,
			key:            key,
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

// decompressBody структура для расшифровки сжатых данных "на ходу"
type decompressBody struct {
	io.ReadCloser
	gzipReader *gzip.Reader
}

// NewDecompressBody инициализация decompressBody
func NewDecompressBody(req *http.Request) (*decompressBody, error) {
	gzipReader, err := gzip.NewReader(req.Body)
	if err != nil {
		return nil, err
	}
	return &decompressBody{
		ReadCloser: req.Body,
		gzipReader: gzipReader,
	}, nil
}

// Read чтение сжатых данных
func (d *decompressBody) Read(p []byte) (n int, err error) {
	return d.gzipReader.Read(p)
}

// Close закрытие всех сущностей decompressBody
func (d *decompressBody) Close() error {
	if err := d.gzipReader.Close(); err != nil {
		return err
	}
	return d.ReadCloser.Close()
}

// compressResponseWriter структура для сжатия данных "на ходу"
// обёртка для ResponseWriter
type compressResponseWriter struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
}

// NewCompressResponseWriter инициализация compressResponseWriter
func NewCompressResponseWriter(w gin.ResponseWriter) (*compressResponseWriter, error) {
	gzipWriter := gzip.NewWriter(w)
	return &compressResponseWriter{
		ResponseWriter: w,
		gzipWriter:     gzipWriter,
	}, nil
}

// Write запись отправляемых данных
func (c *compressResponseWriter) Write(p []byte) (int, error) {
	return c.gzipWriter.Write(p)
}

// WriteHeader записывает заголовки в ResponseWrite
// устанавливает по умолчанию заголовок gzip
func (c *compressResponseWriter) WriteHeader(code int) {
	c.Header().Set("Content-Encoding", "gzip")
	c.ResponseWriter.WriteHeader(code)
}

// Flush делает flush и для gzipWriter и для ResponseWriter
func (c *compressResponseWriter) Flush() {
	c.gzipWriter.Flush()
	c.ResponseWriter.Flush()
}

// Close закрывает gzipWriter
func (c *compressResponseWriter) Close() error {
	return c.gzipWriter.Close()
}

// DataExtraction анализирует пришедшие данные и перенаправляет к нужному обработчику
func DataExtraction() gin.HandlerFunc {
	return func(c *gin.Context) {
		hvCT := c.Request.Header.Values("Content-Encoding")
		var decode bool
		for _, h := range hvCT {
			if strings.Contains(h, "gzip") && !decode {
				dBody, err := NewDecompressBody(c.Request)
				if err != nil {
					c.String(http.StatusInternalServerError, fmt.Sprintf("fail while create decompress request error: %s", err.Error()))
					c.Abort()
					return
				}
				c.Request.Body = dBody
				decode = true
			}
		}

		if strings.Contains(c.Request.URL.Path, "/updates/") {
			c.Next()
		}

		if c.Request.Method == http.MethodPost && strings.Contains(c.Request.URL.Path, "/update/") && strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
			c.Set(METRICTYPE, METRICTYPEJSON)

			c.Next()
		} else {

			c.Set(METRICTYPE, METRICTYPEDEFAULT)

			c.Next()
		}
	}
}

// RespEncode сжимает отправляемые данные
func RespEncode() gin.HandlerFunc {
	return func(c *gin.Context) {
		hvAE := c.Request.Header.Values("Accept-Encoding")
		var encode bool
		for _, h := range hvAE {
			if strings.Contains(h, "gzip") && !encode {
				cW, err := NewCompressResponseWriter(c.Writer)
				if err != nil {
					c.String(http.StatusInternalServerError, fmt.Sprintf("fail while create compress response error: %s", err.Error()))
					c.Abort()
					return
				}
				defer cW.Close()
				defer cW.Flush()
				c.Writer = cW
				encode = true
			}
		}

		c.Next()
	}
}

// PseudoAuth аутентификация
func PseudoAuth(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key != "" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
				c.Abort()
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			receivedHash := c.GetHeader("HashSHA256")
			if receivedHash == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing HashSHA256 header"})
				c.Abort()
				return
			}

			expectedHash := computeHMAC(body, key)
			if receivedHash != expectedHash {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid HMAC"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// WGadd нужно только для того, чтобы обеспечить graceful shutdown
func WGadd (wg *sync.WaitGroup) gin.HandlerFunc {
	wg.Add(1)
	return func(c *gin.Context) {
		c.Next()
	}
}

// computeHMAC высчитывает хэш данных
func computeHMAC(value []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(value)
	return hex.EncodeToString(h.Sum(nil))
}
