package storage

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/Grifonhard/Practicum-metrics/internal/storage/fileio"
)

func New(intervalBackup int, filepathBackup string, restoreFromBackup bool, db psql.StorDB) (*MemStorage, error) {
	var storage MemStorage

	if intervalBackup != 0 {
		storage.backupTicker = time.NewTicker(time.Duration(intervalBackup) * time.Second)
		storage.backupTickerChan = storage.backupTicker.C
	} else {
		storage.backupChan = make(chan struct{})
	}

	var err error

	storage.backupFile, err = fileio.New(filepathBackup, BACKUPFILENAME)
	if err != nil {
		return nil, fmt.Errorf("fail while create/open file: %w", err)
	}

	if db != nil {
		storage.DB = db

		err = db.CreateMetricsTable()
		if err != nil {
			return nil, fmt.Errorf("fail while create metrics table: %w", err)
		}
	} else if restoreFromBackup {
		storage.ItemsGauge, storage.ItemsCounter, err = storage.backupFile.Read()
		if err != nil {
			return nil, fmt.Errorf("fail while read from backup file: %w", err)
		}
	} else {
		storage.ItemsGauge = make(map[string]float64)
		storage.ItemsCounter = make(map[string][]float64)
	}

	return &storage, nil
}

func (ms *MemStorage) Push(metric *Metric) error {
	switch ms.DB {
	case nil:
		ms.mu.Lock()
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
		ms.mu.Unlock()
	default:
		if metric == nil {
			return ErrMetricEmpty
		}
		switch metric.Type {
		case TYPEGAUGE:
			err := ms.DB.PushReplace(metric.Type, metric.Name, metric.Value)
			if err != nil {
				return fmt.Errorf("fail while push gauge to db: %w", err)
			}
		case TYPECOUNTER:
			err := ms.DB.PushAdd(metric.Type, metric.Name, metric.Value)
			if err != nil {
				return fmt.Errorf("fail while push counter to db: %w", err)
			}
		default:
			return ErrMetricTypeUnknown
		}
	}
	if ms.backupChan != nil {
		ms.backupChan <- struct{}{}
	}
	return nil
}

func (ms *MemStorage) Get(metric *Metric) (float64, error) {
	switch ms.DB {
	case nil:
		ms.mu.Lock()
		defer ms.mu.Unlock()
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
	default:
		if metric == nil {
			return 0, ErrMetricEmpty
		}
		switch metric.Type {
		case TYPEGAUGE:
			return ms.DB.GetOneValue(metric.Type, metric.Name)
		case TYPECOUNTER:
			values, err := ms.DB.GetArrayValues(metric.Type, metric.Name)
			if err != nil {
				return 0, fmt.Errorf("error while get value from postgres: %v", err)
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
}

func (ms *MemStorage) List() ([]string, error) {
	var mapGauge map[string]float64
	var mapCounter map[string][]float64
	var err error
	switch ms.DB {
	case nil:
		mapGauge = ms.ItemsGauge
		mapCounter = ms.ItemsCounter
	default:
		mapGauge, mapCounter, err = ms.DB.List(TYPEGAUGE, TYPECOUNTER)
		if err != nil {
			return nil, fmt.Errorf("fail while get list of metrics from db: %w", err)
		}
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	list := make([]string, len(mapGauge)+len(mapCounter))
	var listMu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)
	go listGauge(mapGauge, &list, &wg, &listMu)
	go listCounter(len(mapGauge), mapCounter, &list, &wg, &listMu)

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
		err = ms.backupFile.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("fail whil close file: %v", err))
		}
	}()
	for {
		select {
		case <-ms.backupChan:
			ms.mu.Lock()
			err := ms.backupFile.Write(&fileio.Data{
				ItemsGauge:   ms.ItemsGauge,
				ItemsCounter: ms.ItemsCounter,
			})
			ms.mu.Unlock()
			if err != nil {
				logger.Error(err)
			}
		case <-ms.backupTickerChan:
			ms.mu.Lock()
			err := ms.backupFile.Write(&fileio.Data{
				ItemsGauge:   ms.ItemsGauge,
				ItemsCounter: ms.ItemsCounter,
			})
			ms.mu.Unlock()
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

func listGauge(gauge map[string]float64, list *[]string, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	var i int
	for n, v := range gauge {
		mu.Lock()
		(*list)[i] = fmt.Sprintf("%s: %f", n, v)
		mu.Unlock()
		i++
	}
}

func listCounter(lenGauge int, counter map[string][]float64, list *[]string, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	i := lenGauge
	for n, ves := range counter {
		var values string
		for _, v := range ves {
			values = fmt.Sprintf("%s, %s", values, fmt.Sprintf("%f", v))
		}
		values, _ = strings.CutPrefix(values, ", ")
		mu.Lock()
		(*list)[i] = fmt.Sprintf("%s: %s", n, values)
		mu.Unlock()
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
