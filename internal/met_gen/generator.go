package metgen

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

type MetricsGenerator interface {
	Renew() error
	Collect() (gauge map[string]string, counter map[string]string, err error)
}

type MetGen struct {
	metricsGauge   map[string]float64 //метрики float64
	metricsCounter map[string]int64   //метрики int64
	mu sync.Mutex
}

func New() *MetGen {
	var mg MetGen
	mg.metricsGauge = make(map[string]float64)
	mg.metricsCounter = make(map[string]int64)
	return &mg
}

func (mg *MetGen) Renew() error {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mg.metricsGauge["Alloc"] = float64(memStats.Alloc)
	mg.metricsGauge["BuckHashSys"] = float64(memStats.BuckHashSys)
	mg.metricsGauge["Frees"] = float64(memStats.Frees)
	mg.metricsGauge["GCCPUFraction"] = memStats.GCCPUFraction
	mg.metricsGauge["GCSys"] = float64(memStats.GCSys)
	mg.metricsGauge["HeapAlloc"] = float64(memStats.HeapAlloc)
	mg.metricsGauge["HeapIdle"] = float64(memStats.HeapIdle)
	mg.metricsGauge["HeapInuse"] = float64(memStats.HeapInuse)
	mg.metricsGauge["HeapObjects"] = float64(memStats.HeapObjects)
	mg.metricsGauge["HeapReleased"] = float64(memStats.HeapReleased)
	mg.metricsGauge["HeapSys"] = float64(memStats.HeapSys)
	mg.metricsGauge["LastGC"] = float64(memStats.LastGC)
	mg.metricsGauge["Lookups"] = float64(memStats.Lookups)
	mg.metricsGauge["MCacheInuse"] = float64(memStats.MCacheInuse)
	mg.metricsGauge["MCacheSys"] = float64(memStats.MCacheSys)
	mg.metricsGauge["MSpanInuse"] = float64(memStats.MSpanInuse)
	mg.metricsGauge["MSpanSys"] = float64(memStats.MSpanSys)
	mg.metricsGauge["Mallocs"] = float64(memStats.Mallocs)
	mg.metricsGauge["NextGC"] = float64(memStats.NextGC)
	mg.metricsGauge["NumForcedGC"] = float64(memStats.NumForcedGC)
	mg.metricsGauge["NumGC"] = float64(memStats.NumGC)
	mg.metricsGauge["OtherSys"] = float64(memStats.OtherSys)
	mg.metricsGauge["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	mg.metricsGauge["StackInuse"] = float64(memStats.StackInuse)
	mg.metricsGauge["StackSys"] = float64(memStats.StackSys)
	mg.metricsGauge["Sys"] = float64(memStats.Sys)
	mg.metricsGauge["TotalAlloc"] = float64(memStats.TotalAlloc)
	mg.metricsGauge["RandomValue"] = rand.Float64()

	mg.metricsCounter["PollCount"]++

	return nil
}

func (mg *MetGen) Collect() (map[string]string, map[string]string, error) {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	gg := make(map[string]string)
	cntr := make(map[string]string)
	for k, v := range mg.metricsGauge {
		gg[k] = fmt.Sprint(v)
	}
	for k, v := range mg.metricsCounter {
		cntr[k] = fmt.Sprint(v)
	}
	return gg, cntr, nil
}
