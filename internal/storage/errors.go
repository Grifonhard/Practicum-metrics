package storage

import "errors"

var (
	ErrMetricEmpty              = errors.New("metric is empty")
	ErrMetricTypeUnknown        = errors.New("unknown metrics type")
	ErrMetricNoData             = errors.New("no data for this metric")
	ErrMetricValEmptyField      = errors.New("empty field in metrics")
	ErrMetricValWrongType       = errors.New("wrong type of metrics")
	ErrMetricValValueIsNotFloat = errors.New("value is not float64")
)
