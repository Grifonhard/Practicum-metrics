package psql

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

const METRICSEPARATOR = "///" // METRICSEPARATOR - разделитель для хранения типа и имени в одном поле в БД

// Metric хранит в себе метрику в формате БД
type Metric struct {
	MetricS MetricString
	Value   float64
}

// MetricString хранит в себе тип и имя метрики для автоматического преобразования
type MetricString struct {
	MetricType string
	MetricName string
}

// Value преобразует тип и имя метрики в поле для БД
func (m *MetricString) Value() (driver.Value, error) {
	return m.MetricType + METRICSEPARATOR + m.MetricName, nil
}

// Scan преобразует поле из БД в отдельные поля тип и имя
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
		return ErrConvertProblem
	}
	mFields := strings.Split(mc, METRICSEPARATOR)

	if len(mFields) != 2 {
		return fmt.Errorf("unexpected metric format from db: %s", mc)
	}

	m.MetricType = mFields[0]
	m.MetricName = mFields[1]
	return nil
}
