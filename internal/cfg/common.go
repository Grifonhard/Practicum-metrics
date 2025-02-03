package cfg

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	SECONDS = "s"
	MINUTES = "m"
	HOURS   = "h"
	DAYS    = "d"
)

func parseStrToInt(source *string) (*int, error) {
	if source == nil || *source == "" {
		return nil, nil
	}

	var multiplicator int
	var suf string
	if strings.Contains(*source, "s") {
		multiplicator = 1
		suf = "s"
	} else if strings.Contains(*source, "m") {
		multiplicator = 60
		suf = "m"
	} else if strings.Contains(*source, "h") {
		multiplicator = 3600
		suf = "h"
	} else if strings.Contains(*source, "d") {
		multiplicator = 86400
		suf = "d"
	}

	if multiplicator == 0 {
		return nil, fmt.Errorf("%w %s", ErrWrongTimeFormat, *source)
	}

	cuttedSuff, _ := strings.CutSuffix(*source, suf)

	result, err := strconv.Atoi(cuttedSuff)
	if err != nil {
		return nil, fmt.Errorf("%w %s parse error: %s", ErrWrongTimeFormat, *source, err.Error())
	}

	result *= multiplicator

	return &result, nil
}
