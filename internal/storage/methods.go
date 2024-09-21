package storage

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func New() *MemStorage {
	var storage MemStorage

	storage.ItemsGauge = make(map[string]float64)
	storage.ItemsCounter = make(map[string][]float64)

	return &storage
}

func (ms *MemStorage) Push(metric *Metric) error {
	if metric == nil {
		return errors.New("metric is empty")
	}
	switch metric.Type {
	case TYPEGAUGE:
		ms.ItemsGauge[metric.Name] = metric.Value
		return nil
	case TYPECOUNTER:
		ms.ItemsCounter[metric.Name] = append(ms.ItemsCounter[metric.Name], metric.Value)
		return nil
	default:
		return errors.New("unknown metrics type")
	}
}

func (ms *MemStorage) Get(metric *Metric) (string, error) {
	if metric == nil {
		return "", errors.New("metric is empty")
	}
	switch metric.Type {
	case TYPEGAUGE:
		result, ok := ms.ItemsGauge[metric.Name]
		if !ok {
			return "", errors.New("no data for this metric")
		}
		return fmt.Sprint(result), nil
	case TYPECOUNTER:
		values, ok := ms.ItemsCounter[metric.Name]
		if !ok {
			return "", errors.New("no data for this metric")
		}
		var result float64
		for _, v := range values {
			result += v
		}
		return fmt.Sprint(result), nil
	default:
		return "", errors.New("unknown metrics type")
	}
}

func (ms *MemStorage) List() ([]string, error) {
	list := make([]string, len(ms.ItemsCounter) + len(ms.ItemsCounter))
	var wg sync.WaitGroup
	wg.Add(len(ms.ItemsGauge))
	for n, v := range ms.ItemsGauge {
		go ms.listRoutin(&list, fmt.Sprintf("%s: %f", n, v), wg)
	}	
	for n, ves := range ms.ItemsCounter {
		var values string
		for _, v := range ves {
			values = fmt.Sprintf("%s, %s", values, fmt.Sprintf("%f", v))
		}
		values, _ = strings.CutPrefix(values, ", ")
		wg.Wait()
		list = append(list, fmt.Sprintf("%s: %s", n, values))
	}

	return list, nil
}

func (ms *MemStorage) listRoutin(list *[]string, info string, wg sync.WaitGroup) {
	ms.mu.Lock()
	defer wg.Done()
	defer ms.mu.Unlock()
	(*list) = append((*list), info)
}

func ValidateAndConvert(method, mType, mName, mValue string) (*Metric, error) {
	var result Metric
	var err error

	if method == http.MethodGet {
		mValue = "0"
	}

	if mType == "" || mName == "" || mValue == "" {
		return nil, errors.New("empty field in metrics")
	}

	if mType != TYPEGAUGE && mType != TYPECOUNTER {
		return nil, errors.New("wrong type of metrics")
	} else {
		result.Type = mType
	}
	result.Value, err = strconv.ParseFloat(mValue, 64)
	if err != nil {
		return nil, errors.New("value is not float64")
	}
	result.Name = mName
	return &result, nil
}
