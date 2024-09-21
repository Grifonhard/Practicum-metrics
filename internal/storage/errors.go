package storage

import "errors"

var (
	MetricEmpty              = errors.New("metric is empty")
	MetricTypeUnknown        = errors.New("unknown metrics type")
	MetricNoData             = errors.New("no data for this metric")
	MetricValEmptyField      = errors.New("empty field in metrics")
	MetricValWrongType       = errors.New("wrong type of metrics")
	MetricValValueIsNotFloat = errors.New("value is not float64")
)
