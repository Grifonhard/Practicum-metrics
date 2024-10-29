package metgen

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type MetricsGenerator interface {
	Renew() error
	Collect() (gauge map[string]string, counter map[string]string, err error)
}

type MetGen struct {
	MetricsGauge   map[string]float64 //метрики float64
	MetricsCounter map[string]int64   //метрики int64
	mu sync.RWMutex
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
	input1 := make(chan oneGaugeMetric)
	input2 := make(chan oneGaugeMetric)
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	var closed int

	go getGopsutilMetrics(ctx, input1, errChan)
	go getStandartMetrics(ctx, input2, errChan)

	select {
	case one, ok := <- input1:
		if !ok {
			closed++
		}
		if closed == 2 {
			// это для statictest
			cancel()
			break
		}
		mg.MetricsGauge[one.name] = one.metric
	case one, ok := <- input2:
		if !ok {
			closed++
		}
		if closed == 2 {
			// это для statictest
			cancel()
			break
		}
		mg.MetricsGauge[one.name] = one.metric
	case err := <- errChan:
		cancel()
		<- input1
		<- input2
		return err
	}
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

func getStandartMetrics(ctx context.Context, output chan oneGaugeMetric, errChan chan error) {
	defer close(output)
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

	select {
	case <- ctx.Done():
		return
	default:
		for n, m := range gauge {
			output <- oneGaugeMetric{
				name: n,
				metric: m,
			}
		}
	}
}

func getGopsutilMetrics(ctx context.Context, output chan oneGaugeMetric, errChan chan error) {
	defer close(output)
	gauge := make(map[string]float64)

	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка при получении информации о памяти: %v", err))
		errChan <- err
		return
	}
	if vm != nil {
		gauge["TotalMemory"] = float64(vm.Total)
		gauge["FreeMemory"] = float64(vm.Free)
	}

	cpuUtilization, err := cpu.Percent(1*time.Second, false)
    if err != nil {
        logger.Error(fmt.Sprintf("Ошибка при получении загрузки CPU: %v", err))
		errChan <- err
		return
    }
	if len(cpuUtilization) != 0 {
		gauge["CpuUtilization"] = cpuUtilization[0]
	}

	select {
	case <- ctx.Done():
		return
	default:
		for n, m := range gauge {
			output <- oneGaugeMetric{
				name: n,
				metric: m,
			}
		}
	}
}
