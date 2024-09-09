package storage

const (
	TYPE1 = "gauge"
	TYPE2 = "counter"
)

type Stor interface {
	Push(name, value, typeMetric string) error
	Pop(name string) ([]string, error)
}

type MemStorage struct {
	ItemsGauge   map[string]float64
	ItemsCounter map[string][]float64
}

type Metric struct{
	Type string
	Name string
	Value float64
}