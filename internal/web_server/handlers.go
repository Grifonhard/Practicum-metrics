package webserver

import (
	"fmt"
	"net/http"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/gin-gonic/gin"
)

const (
	PARAMSAMOUNT = 3
	STORAGEKEY   = "storage"
	METRICKEY = "metric"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context){
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

func Update (stor *storage.MemStorage) gin.HandlerFunc{
	return func(c *gin.Context){
		//извлекаем данные из контекста
		itemInter, ok := c.Get(METRICKEY)
		if !ok {
			c.String(http.StatusInternalServerError, "Metric not found in context")
			c.Abort()
			return
		}
		item, ok := itemInter.(*storage.Metric)
		if !ok {
			c.String(http.StatusInternalServerError, "Wrong type of item metric")
			c.Abort()
			return
		}
	
		//сохраняем данные
		err := stor.Push(item)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Fail while push error: %s", err.Error()))
			c.Abort()
			return
		}
		c.Header("Сontent-Length", fmt.Sprint(len("Success")))
		c.Header("Content-Type", "text/plain; charset=utf-8")
	
		c.String(http.StatusOK, "Success")
	}
}

func Get (stor *storage.MemStorage) gin.HandlerFunc{
	return func(c *gin.Context){
		//извлекаем данные из контекста
		itemInter, ok := c.Get(METRICKEY)
		if !ok {
			c.String(http.StatusInternalServerError, "Metric not found in context")
			c.Abort()
			return
		}
		item, ok := itemInter.(*storage.Metric)
		if !ok {
			c.String(http.StatusInternalServerError, "Wrong type of item metric")
			c.Abort()
			return
		}

		//получаем данные
		value, err := stor.Get(item)
		if err != nil && err.Error() == "Metric is empty" {
			c.String(http.StatusInternalServerError, err.Error())
			c.Abort()
			return
		} else if err != nil{
			c.String(http.StatusNotFound, err.Error())
			c.Abort()
			return
		}

		c.Header("Сontent-Length", fmt.Sprint(len(value)))
		c.Header("Content-Type", "text/plain; charset=utf-8")

		c.String(http.StatusOK, value)
	}
}

func List (stor *storage.MemStorage) gin.HandlerFunc{
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