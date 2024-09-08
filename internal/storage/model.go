package storage

const (
	TYPE1 = "gauge"
	TYPE2 = "counter"
)

var MetricNames = []string{
    "Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
    "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
    "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
    "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
    "StackSys", "Sys", "TotalAlloc", "RandomValue", "PollCount",

	"testCounter", "testGauge",
}

type Stor interface {
	Push(name, value, type_metric string) error
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