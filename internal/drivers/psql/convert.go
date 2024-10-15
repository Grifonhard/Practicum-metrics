package psql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

const (
	METRICSEPARATOR = "///"
)

type Metric struct {
	MetricS MetricString
	Value   float64
}

type MetricString struct {
	MetricType string
	MetricName string
}

func (m *MetricString) Value() (driver.Value, error) {
	return m.MetricType + METRICSEPARATOR + m.MetricName, nil
}

func (m *MetricString) Scan(value interface{}) error {
	if value == nil {
		*m = MetricString{}
		return nil
	}

	mr, err := driver.String.ConvertValue(value)
	if err != nil {
		return fmt.Errorf("cannot scan value: %w", err)
	}

	mc, ok := mr.(string)
	if !ok {
		return errors.New("cannot scan value. cannot convert value to string")
	}
	mFields := strings.Split(mc, METRICSEPARATOR)

	if len(mFields) != 2 {
		return fmt.Errorf("unexpected metric format from db: %s", mc)
	}

	m.MetricType = mFields[0]
	m.MetricName = mFields[1]
	return nil
}
