// Модуль отвечает за управление храненением метрик
package storage

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	"github.com/Grifonhard/Practicum-metrics/internal/storage/fileio"
)

// Типы метрик
const (
	TYPEGAUGE      = "gauge"
	TYPECOUNTER    = "counter"
)

// Название файла с бэкапом
const BACKUPFILENAME = "backup"

// Stor внешний интерфейс для хранилища метрик
type Stor interface {
	Push(name, value, typeMetric string) error
	Pop(name string) ([]string, error)
}

// MemStorage структура, которая используется для хранения метрик в оперативной памяти и реализует интерфейс Stor 
type MemStorage struct {
	DB               psql.StorDB
	ItemsGauge       map[string]float64
	ItemsCounter     map[string][]float64
	backupChan       chan struct{}
	backupTickerChan <-chan time.Time
	backupTicker     *time.Ticker
	backupFile       *fileio.File
	mu               sync.Mutex
}

// Metric единичная метрика
type Metric struct {
	Type  string  `json:"type"`
	Name  string  `json:"id"`
	Value float64 `json:"-"`
}

// MarshalJSON кастомная сериализация для Metric 
// необходимо для того, чтобы передавать разные метрики в одном поле json
func (m *Metric) MarshalJSON() ([]byte, error) {
	type MAlias Metric
	switch m.Type {
	case TYPEGAUGE:
		value := m.Value
		return json.Marshal(struct {
			*MAlias
			V *float64 `json:"value,omitempty"`
		}{
			MAlias: (*MAlias)(m),
			V:      &value,
		})
	case TYPECOUNTER:
		dlt := int64(m.Value)
		return json.Marshal(struct {
			*MAlias
			D *int64 `json:"delta,omitempty"`
		}{
			MAlias: (*MAlias)(m),
			D:      &dlt,
		})
	default:
		return nil, ErrMetricTypeUnknown
	}
}

// UnmarshalJSON кастомная десериализация для Metric 
// необходимо для того, чтобы передавать разные метрики в одном поле json
func (m *Metric) UnmarshalJSON(data []byte) error {
	type MAlias Metric
	apiMetric := struct {
		*MAlias
		V *float64 `json:"value,omitempty"`
		D *int64   `json:"delta,omitempty"`
	}{MAlias: (*MAlias)(m)}
	if err := json.Unmarshal(data, &apiMetric); err != nil {
		return err
	}
	switch m.Type {
	case TYPEGAUGE:
		if apiMetric.V == nil {
			return nil
		}
		m.Value = *apiMetric.V
	case TYPECOUNTER:
		if apiMetric.D == nil {
			return nil
		}
		m.Value = float64(*apiMetric.D)
	default:
		return ErrMetricTypeUnknown
	}
	return nil
}
