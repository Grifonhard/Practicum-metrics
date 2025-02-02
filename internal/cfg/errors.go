package cfg

import "errors"

var (
	ErrWrongTimeFormat = errors.New("time format wrong")
	ErrCFGFile = errors.New("problem with cfg file: ")
)
