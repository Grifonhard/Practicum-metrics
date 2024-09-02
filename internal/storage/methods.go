package storage

import "errors"

func New() *MemStorage{
	var store MemStorage

	store.ItemsGauge = make(map[string]string)
	store.ItemsCounter = make(map[string][]string)

	return &store
}

func (ms *MemStorage) Push(name, value, type_metric string) error{
	switch type_metric{
	case "gauge":
		ms.ItemsGauge[name] = value
		return nil
	case "counter":
		ms.ItemsCounter[name] = append(ms.ItemsCounter[name], value)
		return nil
	default:
		return errors.New("Unknown metrics type")
	}
}

func (ms *MemStorage) Pop(name string) ([]string, error){
	return nil, errors.New("Not implemented")
}
