// Модуль используется для передачи данных из агента на сервер хранения метрик
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

	cryptoutils "github.com/Grifonhard/Practicum-metrics/internal/crypto_utils"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
)

// Metrics для сериализации данных из генератора метрик
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Настройки режима отправки данных
const (
	SENDSUBSEQUENCE = "subsequence mode" // для /update
	SENDARRAY       = "array mode"       // для /updates
)

// Настройки повторных попыток отправить данные, если происходят сбои
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

// SendMetric агрегирует и отправляет данные на сервер
func SendMetric(wg *sync.WaitGroup, url string, gen *metgen.MetGen, keyHash, sendMethod string, realIp string) {
	wg.Add(1)
	defer wg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *Metrics)

	// подготовка данных
	var buf bytes.Buffer
	gauge, counter, err := gen.Collect()
	if err != nil {
		logger.Error(fmt.Sprintf("fail collect metrics: %s", err.Error()))
	}
	enc := json.NewEncoder(&buf)
	// используется только для массива итемов /updates
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
			compressed, err := compressBeforeSend(buf.Bytes())
			if err != nil {
				logger.Error(fmt.Sprintf("fail while compress: %s", err.Error()))
				cancel()
				return
			}
			// шифрование, если есть ключ
			var finalBody *bytes.Buffer
			if cryptoutils.PublicKey != nil {
				encrypted, err := cryptoutils.EncryptRSA(compressed.Bytes(), cryptoutils.PublicKey)
				if err != nil {
					logger.Error("error encrypting data: ", err)
					return
				}
				requestBody := []byte(fmt.Sprintf(`{"data":"%s"}`, encrypted))
				finalBody = bytes.NewBuffer(requestBody)
			} else {
				finalBody = compressed
			}
			//подготовка реквеста и клиента
			req, err := http.NewRequest(http.MethodPost, url, finalBody)
			if err != nil {
				logger.Error(fmt.Sprintf("fail while create request: %s", err.Error()))
				cancel()
				return
			}
			if keyHash != "" {
				hmacHash := computeHMAC(compressed.String(), keyHash)
				req.Header.Set("HashSHA256", hmacHash)
			}
			if realIp != "" {
				req.Header.Set("X-Real-IP", realIp)
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

// prepareDataToSend подготовка и отправка данных
// приспособлена для асинхронной работы с функциями отправляющими данные
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

// compressBeforeSend сжатие данных перед отправкой
func compressBeforeSend(b []byte) (compressed *bytes.Buffer, err error) {
	compressed = new(bytes.Buffer)
	writer := gzip.NewWriter(compressed)

	_, err = writer.Write(b)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return compressed, nil
}

// computeHMAC подготовка hmac для отправляемых данных
func computeHMAC(value, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

// SendMetricWithWorkerPool асинхронная подготовка и отправка метрик
func SendMetricWithWorkerPool(wgSig *sync.WaitGroup, url string, gen *metgen.MetGen, keyHash string, rateLimit int, realIp string) {
	wgSig.Add(1)
	defer wgSig.Done()
	collectG := make(chan metgen.OneMetric)
	collectC := make(chan metgen.OneMetric)
	workerChan := make(chan Metrics)
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// запускаем пул воркеров
	for i := 0; i < rateLimit; i++ {
		wg.Add(1)
		go sendWorker(ctx, &wg, url, keyHash, workerChan, errChan, realIp)
	}

	// запуск генераторов
	go gen.CollectGaugeToChan(ctx, collectG, errChan)
	go gen.CollectCounterToChan(ctx, collectC, errChan)

	// собираем данные в канал для воркеров
	go fanIn(ctx, collectG, collectC, workerChan)

	// обработка ошибок
	go func() {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			logger.Error(fmt.Sprintf("fail while sending metrics: %s\n", err.Error()))
			cancel()
			// очищаем каналы чтобы функции передающие данные в момент cancel прервали работу
			// static test не даёт использовать _
			for drop := range collectG {
				logger.Info(fmt.Sprintf("%v dropped", drop))
			}
			for drop := range collectC {
				logger.Info(fmt.Sprintf("%v dropped", drop))
			}
			for drop := range workerChan {
				logger.Info(fmt.Sprintf("%v dropped", drop))
			}
		}
	}()

	wg.Wait()

	// для static
	cancel()

	logger.Info("sending with workers is over")
}

// sendWorker отправка данных на сервер
// предназначена для работы как отдельная горутина
func sendWorker(ctx context.Context, wg *sync.WaitGroup, url, keyHash string, input chan Metrics, errChan chan error, realIp string) {
	defer wg.Done()
	//какого-то хрена заголовок Accept-Encoding gzip устанавливается автоматически в клиенте по умолчанию
	cl := &http.Client{
		Timeout: time.Minute,
	}
	for {
		select {
		case <-ctx.Done():
			return
		case one, ok := <-input:
			if !ok {
				return
			}
			oneMar, err := json.Marshal([]Metrics{one})
			if err != nil {
				errChan <- err
				return
			}
			compressed, err := compressBeforeSend(oneMar)
			if err != nil {
				errChan <- err
				return
			}
			// шифрование, если есть ключ
			var finalBody *bytes.Buffer
			if cryptoutils.PublicKey != nil {
				encrypted, err := cryptoutils.EncryptRSA(compressed.Bytes(), cryptoutils.PublicKey)
				if err != nil {
					logger.Error("error encrypting data: ", err)
					return
				}
				requestBody := []byte(fmt.Sprintf(`{"data":"%s"}`, encrypted))
				finalBody = bytes.NewBuffer(requestBody)
			} else {
				finalBody = compressed
			}
			//подготовка реквеста и клиента
			req, err := http.NewRequest(http.MethodPost, url, finalBody)
			if err != nil {
				errChan <- err
				return
			}
			if keyHash != "" {
				hmacHash := computeHMAC(compressed.String(), keyHash)
				req.Header.Set("HashSHA256", hmacHash)
			}
			if realIp != "" {
				req.Header.Set("X-Real-IP", realIp)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")

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
				errChan <- err
				return
			}
			logger.Info(fmt.Sprintf("one metric send, status: %s\n", resp.Status))
		}
	}
}

// fanIn посредник между продюсерами метрик и воркерами для отправки метрик
func fanIn(ctx context.Context, inputG, inputC chan metgen.OneMetric, output chan Metrics) {
	defer close(output)
	var closed [2]int
	for {
		select {
		case <-ctx.Done():
			return
		case one, ok := <-inputG:
			if !ok {
				closed[0] = 1
			}
			if closed[0] == 1 && closed[1] == 1 {
				return
			}
			var metric Metrics
			metric.ID = one.Name
			metric.MType = storage.TYPEGAUGE
			val := one.Metric
			metric.Value = &val
			output <- metric
		case one, ok := <-inputC:
			if !ok {
				closed[1] = 1
			}
			if closed[0] == 1 && closed[1] == 1 {
				return
			}
			var metric Metrics
			metric.ID = one.Name
			metric.MType = storage.TYPECOUNTER
			dlt := int64(one.Metric)
			metric.Delta = &dlt
			output <- metric
		}
	}
}
