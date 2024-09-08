package storage

import (
	"errors"
	"strconv"
)

func New() *MemStorage {
	var storage MemStorage

	storage.ItemsGauge = make(map[string]string)
	storage.ItemsCounter = make(map[string][]string)

	return &storage
}

func (ms *MemStorage) Push(name, value, typeMetric string) error {
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

func ValidateBeforePush (mType, mName, mValue string) error{
	if mType != TYPE1 && mType != TYPE2{
		return errors.New("Wrong type of metrics")
	}
	_, err := strconv.ParseFloat(mValue, 64)
	if err != nil{
		return errors.New("Value is not float64")
	}
	for _, name := range MetricNames{
		if name == mName{
			return nil
		}
	}
	return errors.New("Wrong metric name")
}