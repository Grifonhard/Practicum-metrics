package webclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
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
			//сжимаем данные
			compressed, err := compressBeforeSend(&buf)
			if err != nil {
				fmt.Printf("fail while compress: %s", err.Error())
				cancel()
				return
			}
			//подготовка реквеста и клиента
			req, err := http.NewRequest(http.MethodPost, url, compressed)
			if err != nil {
				fmt.Printf("fail while create request: %s", err.Error())
				cancel()
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding","gzip")
			//какого-то хрена заголовок Accept-Encoding gzip устанавливается автоматически в клиенте по умолчанию
			cl := &http.Client{
				Timeout: time.Minute,
			}

			resp, err := cl.Do(req)
			if err != nil {
				fmt.Printf("fail while sending metrics: %s", err.Error())
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

func compressBeforeSend(b *bytes.Buffer) (compressed *bytes.Buffer, err error) {
	compressed = new(bytes.Buffer)
	writer := gzip.NewWriter(compressed)
	
	_, err = writer.Write(b.Bytes())
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return compressed, nil
}