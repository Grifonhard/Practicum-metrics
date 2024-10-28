package metgen

import (
	"math/rand"
	"runtime"
	"sync"
)

type MetricsGenerator interface {
	Renew() error
	Collect() (gauge map[string]string, counter map[string]string, err error)
}

type MetGen struct {
	MetricsGauge   map[string]float64 //метрики float64
	MetricsCounter map[string]int64   //метрики int64
	mu sync.Mutex
}

type oneGaugeMetric struct {
	name string
	metric float64
}

func New() *MetGen {
	var mg MetGen
	mg.MetricsGauge = make(map[string]float64)
	mg.MetricsCounter = make(map[string]int64)
	return &mg
}

func (mg *MetGen) Renew() error {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	var wg sync.WaitGroup


	mg.MetricsCounter["PollCount"]++
	wg.Wait()

	return nil
}

func (mg *MetGen) Collect() (map[string]float64, map[string]int64, error) {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	gg := make(map[string]float64)
	cntr := make(map[string]int64)
	for k, v := range mg.MetricsGauge {
		gg[k] = v
	}
	for k, v := range mg.MetricsCounter {
		cntr[k] = v
	}
	return gg, cntr, nil
}

func getStandartMetrics(wg *sync.WaitGroup, output chan <- oneGaugeMetric) {
	defer wg.Done()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	gauge := make(map[string]float64)

	gauge["Alloc"] = float64(memStats.Alloc)
	gauge["BuckHashSys"] = float64(memStats.BuckHashSys)
	gauge["Frees"] = float64(memStats.Frees)
	gauge["GCCPUFraction"] = memStats.GCCPUFraction
	gauge["GCSys"] = float64(memStats.GCSys)
	gauge["HeapAlloc"] = float64(memStats.HeapAlloc)
	gauge["HeapIdle"] = float64(memStats.HeapIdle)
	gauge["HeapInuse"] = float64(memStats.HeapInuse)
	gauge["HeapObjects"] = float64(memStats.HeapObjects)
	gauge["HeapReleased"] = float64(memStats.HeapReleased)
	gauge["HeapSys"] = float64(memStats.HeapSys)
	gauge["LastGC"] = float64(memStats.LastGC)
	gauge["Lookups"] = float64(memStats.Lookups)
	gauge["MCacheInuse"] = float64(memStats.MCacheInuse)
	gauge["MCacheSys"] = float64(memStats.MCacheSys)
	gauge["MSpanInuse"] = float64(memStats.MSpanInuse)
	gauge["MSpanSys"] = float64(memStats.MSpanSys)
	gauge["Mallocs"] = float64(memStats.Mallocs)
	gauge["NextGC"] = float64(memStats.NextGC)
	gauge["NumForcedGC"] = float64(memStats.NumForcedGC)
	gauge["NumGC"] = float64(memStats.NumGC)
	gauge["OtherSys"] = float64(memStats.OtherSys)
	gauge["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	gauge["StackInuse"] = float64(memStats.StackInuse)
	gauge["StackSys"] = float64(memStats.StackSys)
	gauge["Sys"] = float64(memStats.Sys)
	gauge["TotalAlloc"] = float64(memStats.TotalAlloc)
	gauge["RandomValue"] = rand.Float64()

	for n, m := range gauge {
		output <- oneGaugeMetric{
			name: n,
			metric: m,
		}
	}
}

func getGopsutilMetrics(wg *sync.WaitGroup, )