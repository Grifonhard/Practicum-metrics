package webserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
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
				respondWithError(c, http.StatusBadRequest, "validate error", err.Error(), err)
				return
			}

			//сохраняем данные
			err = stor.Push(item)
			if err != nil {
				respondWithError(c, http.StatusInternalServerError, "fail while push error", "fail while push data in db", err)
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
			c.Header("Content-Type", "application/json; charset=utf-8")
			for {
				var item storage.Metric
				err = dec.Decode(&item)
				if err == io.EOF {
					break
				}
				if err != nil {
					respondWithError(c, http.StatusBadRequest, "fail while decode error", "fail unmarshal data", err)
					return
				}

				err = stor.Push(&item)
				if err != nil {
					respondWithError(c, http.StatusInternalServerError, "fail while push error", "fail push data to db", err)
					return
				}

				renewValue, err := stor.Get(&item)
				if err != nil {
					respondWithError(c, http.StatusInternalServerError, "fail while get error", "fail while control renew data", err)
					return
				}

				item.Value = renewValue
				err = enc.Encode(&item)
				if err != nil {
					respondWithError(c, http.StatusInternalServerError, "fail while get error", "fail while marshal data", err)
					return
				}
			}
			c.Writer.WriteHeader(http.StatusOK)
			c.Writer.Write(buf.Bytes())
		default:
			c.String(http.StatusInternalServerError, "wrong metric type in context")
			c.Abort()
			return
		}
	}
}

func Updates(stor *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		c.Header("Content-Type", "application/json; charset=utf-8")
		_, err := buf.ReadFrom(c.Request.Body)		
		if err != nil {
			respondWithError(c, http.StatusBadRequest, "fail while read to buffer error", "fail read data", err)
			return
		}
		var items []storage.Metric
		err = json.Unmarshal(buf.Bytes(), &items)
		if err != nil {
			respondWithError(c, http.StatusBadRequest, "fail while decode error", "fail unmarshal data", err)
			return
		}

		for i, item := range items {
			err = stor.Push(&item)
			if err != nil {
				respondWithError(c, http.StatusInternalServerError, "fail while push error", "fail push data to db", err)
				return
			}

			renewValue, err := stor.Get(&item)
			if err != nil {
				respondWithError(c, http.StatusInternalServerError, "fail while get error", "fail while control renew data", err)
				return
			}

			items[i].Value = renewValue
		}
		c.JSON(http.StatusOK, items)
	}
}

func GetJSON(stor *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		if !strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong content type, only json allow"})
			c.Abort()
			return
		}

		// Для разнообразия принимаем только по одной штуке
		var item storage.Metric

		if err := c.ShouldBindJSON(&item); err != nil {
			respondWithError(c, http.StatusBadRequest, "fail while decode error", "fail unmarshal data", err)
			return
		}

		value, err := stor.Get(&item)
		if err != nil && (errors.Is(err, storage.ErrMetricNoData) || errors.Is(err, psql.ErrNoData)) {
			respondWithError(c, http.StatusNotFound, "fail while get error", "fail get data from db: no data", err)
			return
		} else if err != nil {
			respondWithError(c, http.StatusInternalServerError, "fail while get error", "fail while get data from db", err)
			return
		}

		item.Value = value

		// Логируем заголовки и тело ответа
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
			if err != nil && (errors.Is(err, storage.ErrMetricNoData) || errors.Is(err, psql.ErrNoData)) {
				c.String(http.StatusNotFound, err.Error())
				c.Abort()
				return
			} else if err != nil {
				respondWithError(c, http.StatusInternalServerError, "fail while get error", "data not found", err)
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
			respondWithError(c, http.StatusInternalServerError, "fail while list error", "can't get list of metrics", err)
			return
		}

		c.HTML(http.StatusOK, "list.html", gin.H{
			"Title": "List of metrics",
			"Items": list,
		})
	}
}

func PingDB(db *psql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			respondWithError(c, http.StatusInternalServerError, "fail ping db error", "not pong", err)
			return
		}
		c.String(http.StatusOK, "pong")
	}
}

func respondWithError(c *gin.Context, status int, logMessage string, userMessage string, err error) {
    logger.Error(fmt.Sprintf("%s: %s", logMessage, err.Error()))
    c.JSON(status, gin.H{"error": userMessage})
    c.Abort()
}