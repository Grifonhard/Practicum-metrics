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

func (ms *MemStorage) Push(typeMetric, name string, value float64) error {
	switch typeMetric {
	case TYPE1:
		ms.ItemsGauge[name] = value
		return nil
	case TYPE2:
		ms.ItemsCounter[name] = append(ms.ItemsCounter[name], value)
		return nil
	default:
		return errors.New("Unknown metrics type")
	}
}

func (ms *MemStorage) Pop(name string) ([]string, error) {
	return nil, errors.New("Not implemented")
}

func ValidateBeforePush (mType, mName, mValue string) (float64, error){
	var valueF float64
	if mType != TYPE1 && mType != TYPE2{
		return 0, errors.New("Wrong type of metrics")
	}
	valueF, err := strconv.ParseFloat(mValue, 64)
	if err != nil{
		return 0, errors.New("Value is not float64")
	}
	for _, name := range MetricNames{
		if name == mName{
			return valueF, nil
		}
	}
	return 0, errors.New("Wrong metric name")
}