package webclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

const (
	SENDSUBSEQUENCE = "subsequence mode" // для /update
	SENDARRAY       = "array mode"       // для /updates
)

// если неудачно
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

func SendMetric(url string, gen *metgen.MetGen, keyHash, sendMethod string) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *Metrics)

	// подготовка данных
	var buf bytes.Buffer
	gauge, counter, err := gen.Collect()
	if err != nil {
		logger.Error(fmt.Sprintf("fail collect metrics: %s", err.Error()))
	}
	enc := json.NewEncoder(&buf)
	// используется только для массива итемов
	var items []*Metrics

	go prepareDataToSend(gauge, counter, ch, cancel)
	for {
		select {
		case item := <-ch:
			switch sendMethod {
			case SENDSUBSEQUENCE:
				err = enc.Encode(item)
				if err != nil {
					logger.Error(fmt.Sprintf("fail encode metrics: %s", err.Error()))
				}
			case SENDARRAY:
				items = append(items, item)
			}
		case <-ctx.Done():
			close(ch)
			if sendMethod == SENDARRAY {
				err = enc.Encode(items)
				if err != nil {
					logger.Error(fmt.Sprintf("fail encode metrics: %s", err.Error()))
				}
			}
			//сжимаем данные
			compressed, err := compressBeforeSend(&buf)
			if err != nil {
				logger.Error(fmt.Sprintf("fail while compress: %s", err.Error()))
				cancel()
				return
			}
			//подготовка реквеста и клиента
			req, err := http.NewRequest(http.MethodPost, url, compressed)
			if err != nil {
				logger.Error(fmt.Sprintf("fail while create request: %s", err.Error()))
				cancel()
				return
			}
			if keyHash != "" {
				hmacHash := computeHMAC(compressed.String(), keyHash)
				req.Header.Set("HashSHA256", hmacHash)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")
			//какого-то хрена заголовок Accept-Encoding gzip устанавливается автоматически в клиенте по умолчанию
			cl := &http.Client{
				Timeout: time.Minute,
			}

			var resp *http.Response
			var errCollect []error
			for i := 0; i < MAXRETRIES; i++ {
				resp, err = cl.Do(req)
				if err != nil {
					time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
					errCollect = append(errCollect, err)
					continue
				} else {
					defer resp.Body.Close()
					break
				}
			}
			if errCollect != nil {
				logger.Error(fmt.Sprintf("problem with sending metrics: %s\n", errors.Join(errCollect...).Error()))
			}
			if err != nil {
				logger.Error(fmt.Sprintf("fail while sending metrics: %s\n", err.Error()))
				return
			}

			logger.Info(fmt.Sprintf("success send, status: %s\n", resp.Status))
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

func computeHMAC(value, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}