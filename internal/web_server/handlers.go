package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
			c.Header("Сontent-Length", fmt.Sprint(len("success")))
			c.Header("Content-Type", "text/plain; charset=utf-8")

			c.String(http.StatusOK, "success")
			return
		case METRICTYPEJSON:
			var buf bytes.Buffer
			var err error
			dec := json.NewDecoder(c.Request.Body)
			enc := json.NewEncoder(&buf)
			var valueOld float64
			for {
				var item storage.Metric
				err = dec.Decode(&item)
				if err != nil && err != io.EOF {
					c.String(http.StatusBadRequest, fmt.Sprintf("fail while decode json: %s", err.Error()))
					c.Abort()
					return
				} else if err != nil && err == io.EOF {
					break
				}

				if item.Type == storage.TYPECOUNTER {
					valueOld, err = getAndConvert(stor, &item)
					if err != nil {
						c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
						c.Abort()
						return
					}
				}

				err = stor.Push(&item)
				if err != nil {
					c.String(http.StatusInternalServerError, fmt.Sprintf("fail while push error: %s", err.Error()))
					c.Abort()
					return
				}

				renewValue, err := getAndConvert(stor, &item)
				if err != nil {
					c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
					c.Abort()
					return
				}

				switch item.Type {
				case storage.TYPECOUNTER:
					item.Value = renewValue - valueOld
					err = enc.Encode(item)
					if err != nil {
						c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err.Error()))
						c.Abort()
						return
					}
				case storage.TYPEGAUGE:
					item.Value = renewValue
					err = enc.Encode(item)
					if err != nil {
						c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err.Error()))
						c.Abort()
						return
					}
				}
			}
			c.Header("Сontent-Length", fmt.Sprint(buf.Len()))
			c.Data(http.StatusOK, "application/json", buf.Bytes())
		default:
			c.String(http.StatusInternalServerError, "wrong metric type in context")
			c.Abort()
			return
		}
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

			c.Header("Сontent-Length", fmt.Sprint(len(value)))
			c.Header("Content-Type", "text/plain; charset=utf-8")

			c.String(http.StatusOK, value)
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

func getAndConvert(stor *storage.MemStorage, metric *storage.Metric) (float64, error) {
	valueS, err := stor.Get(metric)
	if err != nil {
		return 0, fmt.Errorf("fail while get error: %v", err)
	}
	fl, err := strconv.ParseFloat(valueS, 64)
	if err != nil {
		return 0, fmt.Errorf("fail while convert to float error: %v", err)
	}
	return fl, nil
}
