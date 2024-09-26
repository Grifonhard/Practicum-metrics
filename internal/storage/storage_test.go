package storage

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T) {
	stor := New()

	metrics := []Metric{
		{
			Type:  "gauge",
			Name:  "name1",
			Value: 1.12,
		},
		{
			Type:  "gauge",
			Name:  "name1",
			Value: 2.24,
		},
		{
			Type:  "counter",
			Name:  "name2",
			Value: 3.36,
		},
		{
			Type:  "counter",
			Name:  "name2",
			Value: 4.48,
		},
	}

	err := stor.Push(&metrics[0])
	assert.NoError(t, err)
	assert.Equal(t, stor.ItemsGauge[metrics[0].Name], metrics[0].Value)
	err = stor.Push(&metrics[1])
	assert.Equal(t, stor.ItemsGauge[metrics[0].Name], metrics[1].Value)
	assert.NoError(t, err)

	err = stor.Push(&metrics[2])
	assert.NoError(t, err)
	err = stor.Push(&metrics[3])
	assert.NoError(t, err)

	assert.Equal(t, stor.ItemsCounter[metrics[3].Name], []float64{metrics[2].Value, metrics[3].Value})
}

func TestMar(t *testing.T) {
	var item Metric
	item.Name = "test"
	item.Type = TYPEGAUGE
	item.Value = 1.11

	jn, err := json.Marshal(&item)
	assert.NoError(t, err)

	fmt.Println(string(jn))
}

func TestUnmar(t *testing.T) {
	var item Metric
	err := json.Unmarshal([]byte(`{"id":"PollCount","type":"counter"}`), &item)
	assert.NoError(t, err)
	fmt.Println(item)
}