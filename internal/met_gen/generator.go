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

func New() *MetGen {
	var mg MetGen
	mg.MetricsGauge = make(map[string]float64)
	mg.MetricsCounter = make(map[string]int64)
	return &mg
}

func (mg *MetGen) Renew() error {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mg.MetricsGauge["Alloc"] = float64(memStats.Alloc)
	mg.MetricsGauge["BuckHashSys"] = float64(memStats.BuckHashSys)
	mg.MetricsGauge["Frees"] = float64(memStats.Frees)
	mg.MetricsGauge["GCCPUFraction"] = memStats.GCCPUFraction
	mg.MetricsGauge["GCSys"] = float64(memStats.GCSys)
	mg.MetricsGauge["HeapAlloc"] = float64(memStats.HeapAlloc)
	mg.MetricsGauge["HeapIdle"] = float64(memStats.HeapIdle)
	mg.MetricsGauge["HeapInuse"] = float64(memStats.HeapInuse)
	mg.MetricsGauge["HeapObjects"] = float64(memStats.HeapObjects)
	mg.MetricsGauge["HeapReleased"] = float64(memStats.HeapReleased)
	mg.MetricsGauge["HeapSys"] = float64(memStats.HeapSys)
	mg.MetricsGauge["LastGC"] = float64(memStats.LastGC)
	mg.MetricsGauge["Lookups"] = float64(memStats.Lookups)
	mg.MetricsGauge["MCacheInuse"] = float64(memStats.MCacheInuse)
	mg.MetricsGauge["MCacheSys"] = float64(memStats.MCacheSys)
	mg.MetricsGauge["MSpanInuse"] = float64(memStats.MSpanInuse)
	mg.MetricsGauge["MSpanSys"] = float64(memStats.MSpanSys)
	mg.MetricsGauge["Mallocs"] = float64(memStats.Mallocs)
	mg.MetricsGauge["NextGC"] = float64(memStats.NextGC)
	mg.MetricsGauge["NumForcedGC"] = float64(memStats.NumForcedGC)
	mg.MetricsGauge["NumGC"] = float64(memStats.NumGC)
	mg.MetricsGauge["OtherSys"] = float64(memStats.OtherSys)
	mg.MetricsGauge["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	mg.MetricsGauge["StackInuse"] = float64(memStats.StackInuse)
	mg.MetricsGauge["StackSys"] = float64(memStats.StackSys)
	mg.MetricsGauge["Sys"] = float64(memStats.Sys)
	mg.MetricsGauge["TotalAlloc"] = float64(memStats.TotalAlloc)
	mg.MetricsGauge["RandomValue"] = rand.Float64()

	mg.MetricsCounter["PollCount"]++

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
