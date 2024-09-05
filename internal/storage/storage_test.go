package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T){
	stor := New()

	err := stor.Push("name1", "value1", "gauge")
	assert.NoError(t, err)
	assert.Equal(t, stor.ItemsGauge["name1"], "value1")
	err = stor.Push("name1", "value2", "gauge")
	assert.Equal(t, stor.ItemsGauge["name1"], "value2")
	assert.NoError(t, err)

	err = stor.Push("name2", "value3", "counter")
	assert.NoError(t, err)
	err = stor.Push("name2", "value4", "counter")
	assert.NoError(t, err)
	
	assert.Equal(t, stor.ItemsCounter["name2"], []string{"value3","value4"})	
}