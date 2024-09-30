package storage

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage/fileio"
)

func New(interval int, filepath string, restore bool) (*MemStorage, error) {
	var storage MemStorage

	if interval != 0 {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		storage.backupTicker = ticker
	} else {
		storage.backupChan = make(chan time.Time)
	}

	storage.backupFile = fileio.New(filepath, BACKUPFILENAME)

	var err error
	if restore {
		storage.ItemsGauge, storage.ItemsCounter, err = storage.backupFile.Read()
		if err != nil {
			return nil, err
		}
	} else {
		storage.ItemsGauge = make(map[string]float64)
		storage.ItemsCounter = make(map[string][]float64)
	}

	return &storage, nil
}

func (ms *MemStorage) Push(metric *Metric) error {
	if metric == nil {
		return ErrMetricEmpty
	}
	switch metric.Type {
	case TYPEGAUGE:
		ms.ItemsGauge[metric.Name] = metric.Value
	case TYPECOUNTER:
		ms.ItemsCounter[metric.Name] = append(ms.ItemsCounter[metric.Name], metric.Value)
	default:
		return ErrMetricTypeUnknown
	}
	if ms.backupChan != nil {
		ms.backupChan <- time.Now()
	}
	return nil
}

func (ms *MemStorage) Get(metric *Metric) (float64, error) {
	if metric == nil {
		return 0, ErrMetricEmpty
	}
	switch metric.Type {
	case TYPEGAUGE:
		result, ok := ms.ItemsGauge[metric.Name]
		if !ok {
			return 0, ErrMetricNoData
		}
		return result, nil
	case TYPECOUNTER:
		values, ok := ms.ItemsCounter[metric.Name]
		if !ok {
			return 0, ErrMetricNoData
		}
		var result float64
		for _, v := range values {
			result += v
		}
		return result, nil
	default:
		return 0, ErrMetricTypeUnknown
	}
}

func (ms *MemStorage) List() ([]string, error) {
	list := make([]string, len(ms.ItemsCounter)+len(ms.ItemsCounter))
	var wg sync.WaitGroup
	wg.Add(2)
	go ms.listGauge(&list, &wg)
	go ms.listCounter(&list, &wg)

	wg.Wait()

	return list, nil
}

func (ms *MemStorage) BackupLoop() {
	defer func() {
		ms.mu.Lock()
		err := ms.backupFile.Write(&fileio.Data{
			ItemsGauge:   ms.ItemsGauge,
			ItemsCounter: ms.ItemsCounter,
		})
		ms.mu.Unlock()
		if err != nil {
			logger.Error(err)
		}
		if ms.backupTicker != nil {
			ms.backupTicker.Stop()
		}
		if ms.backupChan != nil {
			close(ms.backupChan)
		}
	}()
	for {
		select {
		case _ = <-ms.backupChan:
			ms.mu.Lock()
			ms.backupFile.Write(&fileio.Data{
				ItemsGauge:   ms.ItemsGauge,
				ItemsCounter: ms.ItemsCounter,
			})
			ms.mu.Unlock()
		case _ = <- ms.backupTicker.C:
			ms.mu.Lock()
			ms.backupFile.Write(&fileio.Data{
				ItemsGauge:   ms.ItemsGauge,
				ItemsCounter: ms.ItemsCounter,
			})
			ms.mu.Unlock()
		}
	}
}

func (ms *MemStorage) listGauge(list *[]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var i int
	for n, v := range ms.ItemsGauge {
		ms.mu.Lock()
		(*list)[i] = fmt.Sprintf("%s: %f", n, v)
		ms.mu.Unlock()
		i++
	}
}

func (ms *MemStorage) listCounter(list *[]string, wg *sync.WaitGroup) {
	defer wg.Done()
	i := len(ms.ItemsGauge)
	for n, ves := range ms.ItemsCounter {
		var values string
		for _, v := range ves {
			values = fmt.Sprintf("%s, %s", values, fmt.Sprintf("%f", v))
		}
		values, _ = strings.CutPrefix(values, ", ")
		(*list)[i] = fmt.Sprintf("%s: %s", n, values)
		i++
	}
}

func ValidateAndConvert(method, mType, mName, mValue string) (*Metric, error) {
	var result Metric
	var err error

	if method == http.MethodGet {
		mValue = "0"
	}

	if mType == "" || mName == "" || mValue == "" {
		return nil, ErrMetricValEmptyField
	}

	if mType != TYPEGAUGE && mType != TYPECOUNTER {
		return nil, ErrMetricValWrongType
	} else {
		result.Type = mType
	}
	result.Value, err = strconv.ParseFloat(mValue, 64)
	if err != nil {
		return nil, ErrMetricValValueIsNotFloat
	}
	result.Name = mName
	return &result, nil
}
