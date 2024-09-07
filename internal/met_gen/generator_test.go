package metgen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenew(t *testing.T){
	gen := New()

	err := gen.Renew()
	assert.NoError(t, err)

	assert.Equal(t, 28, len(gen.metricsGauge), fmt.Sprintf("gauge map len: %d", len(gen.metricsGauge)))

	assert.Equal(t, 1, len(gen.metricsCounter), fmt.Sprintf("counter map len: %d", len(gen.metricsGauge)))
}

func TestCollect(t *testing.T){
	gen := New()

	err := gen.Renew()
	assert.NoError(t, err)

	mapG, mapC, err := gen.Collect()
	assert.NoError(t, err)

	assert.Equal(t, 28, len(mapG), fmt.Sprintf("gauge map len: %d", len(mapG)))

	assert.Equal(t, 1, len(mapC), fmt.Sprintf("counter map len: %d", len(mapC)))
}