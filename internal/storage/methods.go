package storage

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func New() *MemStorage {
	var storage MemStorage

	storage.ItemsGauge = make(map[string]float64)
	storage.ItemsCounter = make(map[string][]float64)

	return &storage
}

func (ms *MemStorage) Push(metric *Metric) error {
	if metric == nil{
		return errors.New("Metric is empty")
	}
	switch metric.Type {
	case TYPE1:
		ms.ItemsGauge[metric.Name] = metric.Value
		return nil
	case TYPE2:
		ms.ItemsCounter[metric.Name] = append(ms.ItemsCounter[metric.Name], metric.Value)
		return nil
	default:
		return errors.New("Unknown metrics type")
	}
}

func (ms *MemStorage) Get(metric *Metric) (string, error) {
	if metric == nil{
		return "", errors.New("Metric is empty")
	}
	switch metric.Type {
	case TYPE1:
		result, ok := ms.ItemsGauge[metric.Name]
		if !ok{
			return "", errors.New("No data for this metric")
		}
		return fmt.Sprint(result), nil
	case TYPE2:
		result, ok := ms.ItemsCounter[metric.Name]
		if !ok{
			return "", errors.New("No data for this metric")
		}
		last := len(result) - 1
		return fmt.Sprint(result[last]), nil
	default:
		return "", errors.New("Unknown metrics type")
	}
}

func (ms *MemStorage) List() ([]string, error) {
	var list []string
	for n, v := range ms.ItemsGauge{
		list = append(list, fmt.Sprintf("%s: %f", n, v))
	}
	for n, ves := range ms.ItemsCounter{
		var values string
		for _, v := range ves{
			values = fmt.Sprintf("%s, %s", values, fmt.Sprintf("%f", v))
		}
		values, _ = strings.CutPrefix(values, ", ")
		list = append(list, fmt.Sprintf("%s: %s", n, values))
	}
	return list, nil
}

func ValidateAndConvert (method, mType, mName, mValue string) (*Metric, error){
	var result Metric
	var err error

	if method == http.MethodGet {
		mValue = "0"
	}

	if mType == "" || mName == "" || mValue == "" {
		return nil, errors.New("Empty field in metrics")
	}
	
	if mType != TYPE1 && mType != TYPE2{
		return nil, errors.New("Wrong type of metrics")
	} else {
		result.Type = mType
	}
	result.Value, err = strconv.ParseFloat(mValue, 64)
	if err != nil{
		return nil, errors.New("Value is not float64")
	}
	result.Name = mName
	return &result, nil
}