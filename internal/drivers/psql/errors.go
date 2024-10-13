package psql

import "errors"

var (
	ErrNoData               = errors.New("no data in db")
	ErrUnexpectedMetricType = errors.New("unexpected metric type from db")
)
