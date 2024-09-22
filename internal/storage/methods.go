package storage

import (
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
		return ErrMetricEmpty
	}
	switch metric.Type {
	case TYPEGAUGE:
		ms.ItemsGauge[metric.Name] = metric.Value
		return nil
	case TYPECOUNTER:
		ms.ItemsCounter[metric.Name] = append(ms.ItemsCounter[metric.Name], metric.Value)
		return nil
	default:
		return ErrMetricTypeUnknown
	}
}

func (ms *MemStorage) Get(metric *Metric) (string, error) {
	if metric == nil {
		return "", ErrMetricEmpty
	}
	switch metric.Type {
	case TYPEGAUGE:
		result, ok := ms.ItemsGauge[metric.Name]
		if !ok {
			return "", ErrMetricNoData
		}
		return fmt.Sprint(result), nil
	case TYPECOUNTER:
		values, ok := ms.ItemsCounter[metric.Name]
		if !ok {
			return "", ErrMetricNoData
		}
		var result float64
		for _, v := range values {
			result += v
		}
		return fmt.Sprint(result), nil
	default:
		return "", ErrMetricTypeUnknown
	}
}

func (ms *MemStorage) List() ([]string, error) {
	list := make([]string, len(ms.ItemsCounter)+len(ms.ItemsCounter))
	var wg sync.WaitGroup
	wg.Add(2)
	go ms.listGauge(&list, &wg)
	go ms.listCounter(&list, &wg)

	wg.Wait()

	return list, nil
}

func (ms *MemStorage) listGauge(list *[]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var i int
	for n, v := range ms.ItemsGauge {
		ms.mu.Lock()
		(*list)[i] = fmt.Sprintf("%s: %f", n, v)
		ms.mu.Unlock()
		i++
	}
}

func (ms *MemStorage) listCounter(list *[]string, wg *sync.WaitGroup) {
	defer wg.Done()
	i := len(ms.ItemsGauge)
	for n, ves := range ms.ItemsCounter {
		var values string
		for _, v := range ves {
			values = fmt.Sprintf("%s, %s", values, fmt.Sprintf("%f", v))
		}
		values, _ = strings.CutPrefix(values, ", ")
		(*list)[i] = fmt.Sprintf("%s: %s", n, values)
		i++
	}
}

func ValidateAndConvert(method, mType, mName, mValue string) (*Metric, error) {
	var result Metric
	var err error

	if method == http.MethodGet {
		mValue = "0"
	}

	if mType == "" || mName == "" || mValue == "" {
		return nil, ErrMetricValEmptyField
	}

	if mType != TYPEGAUGE && mType != TYPECOUNTER {
		return nil, ErrMetricValWrongType
	} else {
		result.Type = mType
	}
	result.Value, err = strconv.ParseFloat(mValue, 64)
	if err != nil {
		return nil, ErrMetricValValueIsNotFloat
	}
	result.Name = mName
	return &result, nil
}
