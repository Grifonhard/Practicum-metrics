package storage

type Stor interface{
	Push(name, value, type_metric string) error
	Pop(name string) ([]string, error)
}

type MemStorage struct{
	ItemsGauge	map[string]string
	ItemsCounter map[string][]string
}