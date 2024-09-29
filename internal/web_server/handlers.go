package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/gin-gonic/gin"
)

const (
	PARAMSAMOUNT      = 3
	STORAGEKEY        = "storage"
	METRICTYPE        = "metric_type"
	METRICTYPEJSON    = "json"
	METRICTYPEDEFAULT = "default"
)

func Update(stor *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		mType, ok := c.Get(METRICTYPE)
		if !ok {
			c.String(http.StatusInternalServerError, "metric type not found in context")
			c.Abort()
			return
		}

		switch mType {
		case METRICTYPEDEFAULT:
			item, err := storage.ValidateAndConvert(c.Request.Method, c.Param("type"), c.Param("name"), c.Param("value"))
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				c.Abort()
				return
			}

			//сохраняем данные
			err = stor.Push(item)
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("fail while push error: %s", err.Error()))
				c.Abort()
				return
			}
			c.Header("Content-Length", fmt.Sprint(len("success")))
			c.Header("Content-Type", "text/plain; charset=utf-8")

			c.String(http.StatusOK, "success")
			return
		case METRICTYPEJSON:
			var buf bytes.Buffer
			var err error
			dec := json.NewDecoder(c.Request.Body)
			enc := json.NewEncoder(&buf)
			for {
				var item storage.Metric
				err = dec.Decode(&item)
				if err == io.EOF {
					break
				}
				if err != nil {
					c.Header("Content-Type", "application/json; charset=utf-8")
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					c.Abort()
					return
				}

				err = stor.Push(&item)
				if err != nil {
					c.Header("Content-Type", "application/json; charset=utf-8")
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					c.Abort()
					return
				}

				renewValue, err := stor.Get(&item)
				if err != nil {
					c.Header("Content-Type", "application/json; charset=utf-8")
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					c.Abort()
					return
				}

				item.Value = renewValue
				err = enc.Encode(&item)
				if err != nil {
					c.Header("Content-Type", "application/json; charset=utf-8")
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					c.Abort()
					return
				}
			}
			c.Header("Content-Length", fmt.Sprint(buf.Len()))
			c.Data(http.StatusOK, "application/json; charset=utf-8", buf.Bytes())
		default:
			c.String(http.StatusInternalServerError, "wrong metric type in context")
			c.Abort()
			return
		}
	}
}

func GetJSON(stor *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong content type, only json allow"})
			c.Abort()
			return
		}

		// Для разнообразия принимаем только по одной штуке
		var item storage.Metric

		if err := c.ShouldBindJSON(&item); err != nil {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		value, err := stor.Get(&item)
		if err != nil && err == storage.ErrMetricNoData {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			c.Abort()
			return
		} else if err != nil {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		item.Value = value

		// Логируем заголовки и тело ответа
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, &item)
	}
}

func Get(stor *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		mType, ok := c.Get(METRICTYPE)
		if !ok {
			c.String(http.StatusInternalServerError, "metric type not found in context")
			c.Abort()
			return
		}

		switch mType {
		case METRICTYPEDEFAULT:
			item, err := storage.ValidateAndConvert(c.Request.Method, c.Param("type"), c.Param("name"), c.Param("value"))
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				c.Abort()
				return
			}

			//получаем данные
			value, err := stor.Get(item)
			if err != nil && err == storage.ErrMetricEmpty {
				c.String(http.StatusInternalServerError, err.Error())
				c.Abort()
				return
			} else if err != nil {
				c.String(http.StatusNotFound, err.Error())
				c.Abort()
				return
			}

			c.Header("Сontent-Length", fmt.Sprint(len(fmt.Sprint(value))))
			c.Header("Content-Type", "text/plain; charset=utf-8")

			c.String(http.StatusOK, fmt.Sprint(value))
			return
		default:
			c.String(http.StatusInternalServerError, "wrong metric type in context")
			c.Abort()
			return
		}
	}
}

func List(stor *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		list, err := stor.List()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			c.Abort()
			return
		}

		c.HTML(http.StatusOK, "list.html", gin.H{
			"Title": "List of metrics",
			"Items": list,
		})
	}
}
