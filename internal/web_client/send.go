package webclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func SendMetric(url string, gen *metgen.MetGen) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *Metrics)

	//подготовка данных
	var buf bytes.Buffer
	gauge, counter, err := gen.Collect()
	if err != nil {
		fmt.Printf("fail collect metrics: %s", err.Error())
	}
	enc := json.NewEncoder(&buf)

	//подготовка реквеста и клиента
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		fmt.Printf("fail while create request: %s", err.Error())
		cancel()
		return
	}
	req.Header.Set("Content-Type", "application/json")
	cl := &http.Client{}
	cl.Timeout = time.Minute

	go prepareDataToSend(gauge, counter, ch, cancel)
	for {
		select {
		case item := <-ch:
			err = enc.Encode(item)
			if err != nil {
				fmt.Printf("fail encode metrics: %s", err.Error())
			}
		case <-ctx.Done():
			close(ch)
			if err != nil {
				log.Fatal(err)
			}
			resp, err := cl.Do(req)
			if err != nil {
				fmt.Printf("Fail while sending metrics: %s", err.Error())
				return
			}
			defer resp.Body.Close()

			fmt.Printf("success send, status: %s\n", resp.Status)
			return
		}
	}
}

func prepareDataToSend(g map[string]float64, c map[string]int64, ch chan *Metrics, cancel context.CancelFunc) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for k, v := range g {
			var metric Metrics
			metric.ID = k
			metric.MType = storage.TYPEGAUGE
			val := v
			metric.Value = &val
			ch <- &metric
		}
	}()

	go func() {
		defer wg.Done()
		for k, v := range c {
			var metric Metrics
			metric.ID = k
			metric.MType = storage.TYPECOUNTER
			dlt := v
			metric.Delta = &dlt
			ch <- &metric
		}
	}()
	wg.Wait()	
	cancel()
}
