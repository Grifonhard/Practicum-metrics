package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T){
	stor := New()

	err := stor.Push("name1", "gauge", 1.12)
	assert.NoError(t, err)
	assert.Equal(t, stor.ItemsGauge["name1"], 1.12)
	err = stor.Push("name1", "gauge", 2.24)
	assert.Equal(t, stor.ItemsGauge["name1"], 2.24)
	assert.NoError(t, err)

	err = stor.Push("name2", "counter", 3.36)
	assert.NoError(t, err)
	err = stor.Push("name2", "counter", 4.48)
	assert.NoError(t, err)
	
	assert.Equal(t, stor.ItemsCounter["name2"], []float64{3.36, 4.48})	
}