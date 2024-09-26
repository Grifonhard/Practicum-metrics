package storage

import (
	"encoding/json"
	"sync"
)

const (
	TYPEGAUGE   = "gauge"
	TYPECOUNTER = "counter"
)

type Stor interface {
	Push(name, value, typeMetric string) error
	Pop(name string) ([]string, error)
}

type MemStorage struct {
	ItemsGauge   map[string]float64
	ItemsCounter map[string][]float64
	mu           sync.Mutex
}

type Metric struct {
	Type  string `json:"id"`
	Name  string `json:"type"`
	Value float64 `json:"-"`
}

func (m *Metric) MarshalJSON() ([]byte, error) {
	mAlias := *m
	switch m.Type {
	case TYPEGAUGE:
		value := m.Value
		return json.Marshal(struct {
			Mtrc *Metric
			V    *float64 `json:"value,omitempty"`
		}{
			Mtrc: &mAlias,
			V:    &value,
		})
	case TYPECOUNTER:
		dlt := int64(m.Value)
		return json.Marshal(struct {
			Mtrc *Metric
			D    *int64 `json:"delta,omitempty"`
		}{
			Mtrc: m,
			D:    &dlt,
		})
	default:
		return nil, ErrMetricTypeUnknown
	}
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	mAlias := *m
	apiMetric := struct {
		mtrc *Metric
		V    *float64 `json:"value,omitempty"`
		D    *int64   `json:"delta,omitempty"`
	}{mtrc: &mAlias}
	if err := json.Unmarshal(data, &apiMetric); err != nil {
		return err
	}
	*m = mAlias
	switch apiMetric.mtrc.Type {
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
