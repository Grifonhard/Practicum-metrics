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
	mu             sync.RWMutex
}

type OneMetric struct {
	Name   string
	Metric float64
}

func New() *MetGen {
	var mg MetGen
	mg.MetricsGauge = make(map[string]float64)
	mg.MetricsCounter = make(map[string]int64)
	return &mg
}

// для тестов
var getGopsutilMetricsFunc = getGopsutilMetrics
var getStandartMetricsFunc = getStandartMetrics

func (mg *MetGen) Renew() error {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	input1 := make(chan OneMetric)
	input2 := make(chan OneMetric)
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	var closed [2]int

	go getGopsutilMetricsFunc(ctx, input1, errChan)
	go getStandartMetricsFunc(ctx, input2, errChan)

loop:
	for {
		select {
		case one, ok := <-input1:
			if !ok {
				closed[0] = 1
			}
			if closed[0] == 1 && closed[1] == 1 {
				break loop
			}
			mg.MetricsGauge[one.Name] = one.Metric
		case one, ok := <-input2:
			if !ok {
				closed[1] = 1
			}
			if closed[0] == 1 && closed[1] == 1 {
				break loop
			}
			mg.MetricsGauge[one.Name] = one.Metric
		case err := <-errChan:
			cancel()
			// очищаем каналы чтобы функции передающие данные в момент cancel прервали работу
			// static test не даёт использовать _
			for drop := range input1 {
				logger.Info(fmt.Sprintf("%v dropped", drop))
			}
			for drop := range input2 {
				logger.Info(fmt.Sprintf("%v dropped", drop))
			}
			for drop := range errChan {
				logger.Info(fmt.Sprintf("%v dropped", drop))
			}
			return err
		}
	}
	mg.MetricsCounter["PollCount"]++

	cancel()

	return nil
}

func (mg *MetGen) Collect() (map[string]float64, map[string]int64, error) {
	mg.mu.RLock()
	defer mg.mu.RUnlock()
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

func (mg *MetGen) CollectGaugeToChan(ctx context.Context, output chan OneMetric, errChan chan error) {
	defer close(output)
	mg.mu.RLock()
	for k, v := range mg.MetricsGauge {
		mg.mu.RUnlock()
		select {
		case <-ctx.Done():
			return
		default:
			output <- OneMetric{
				Name:   k,
				Metric: v,
			}
		}
		mg.mu.RLock()
	}
	mg.mu.RUnlock()
}

func (mg *MetGen) CollectCounterToChan(ctx context.Context, output chan OneMetric, errChan chan error) {
	defer close(output)
	mg.mu.RLock()
	for k, v := range mg.MetricsCounter {
		mg.mu.RUnlock()
		select {
		case <-ctx.Done():
			return
		default:
			output <- OneMetric{
				Name:   k,
				Metric: float64(v),
			}
		}
		mg.mu.RLock()
	}
	mg.mu.RUnlock()
}

func getStandartMetrics(ctx context.Context, output chan OneMetric, errChan chan error) {
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

	for n, m := range gauge {
		select {
		case <-ctx.Done():
			return
		default:
			output <- OneMetric{
				Name:   n,
				Metric: m,
			}
		}
	}
}

func getGopsutilMetrics(ctx context.Context, output chan OneMetric, errChan chan error) {
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

	for n, m := range gauge {
		select {
		case <-ctx.Done():
			return
		default:
			output <- OneMetric{
				Name:   n,
				Metric: m,
			}
		}
	}
}
