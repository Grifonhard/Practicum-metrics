package storage

import (
	"errors"
	"strconv"
)

func New() *MemStorage {
	var storage MemStorage

	storage.ItemsGauge = make(map[string]float64)
	storage.ItemsCounter = make(map[string][]float64)

	return &storage
}

func (ms *MemStorage) Push(metric *Metric) error {
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

func (ms *MemStorage) Pop(name string) ([]string, error) {
	return nil, errors.New("Not implemented")
}

func ValidateAndConvert (mType, mName, mValue string) (*Metric, error){
	var result Metric
	var err error
	
	if mType != TYPE1 && mType != TYPE2{
		return nil, errors.New("Wrong type of metrics")
	} else {
		result.Type = mType
	}
	result.Value, err = strconv.ParseFloat(mValue, 64)
	if mValue == "" || err != nil{
		return nil, errors.New("Value is not float64")
	}
	result.Name = mName
	for _, name := range MetricNames{
		if name == mName{
			return &result, nil
		}
	}
	return nil, errors.New("Wrong metric name")
}