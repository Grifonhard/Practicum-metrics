package metgen_test

import (
	"context"
	"fmt"

	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
)

// ExampleMetGen_CollectGaugeToChan демонстрирует, как воспользоваться методом CollectGaugeToChan 
// для получения метрик типа Gauge через канал.
func ExampleMetGen_CollectGaugeToChan() {
	// Создаём контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаём экземпляр MetGen
	mg := metgen.New()

	// Добавим несколько тестовых метрик
	mg.MetricsGauge["Alloc"] = 100
	mg.MetricsGauge["HeapAlloc"] = 200

	// Создаём каналы для вывода метрик и ошибок
	outputChan := make(chan metgen.OneMetric)
	errChan := make(chan error)

	// Вызываем CollectGaugeToChan в отдельной горутине
	go mg.CollectGaugeToChan(ctx, outputChan, errChan)

	// Читаем данные из канала
	for metric := range outputChan {
		fmt.Printf("Gauge Metric: %s = %f\n", metric.Name, metric.Metric)
	}

	// Читаем возможные ошибки (если они случились)
	select {
	case err := <-errChan:
		fmt.Printf("Error: %v\n", err)
	default:
		// Ошибок нет
	}

	// Output:
	// Gauge Metric: Alloc = 100.000000
	// Gauge Metric: HeapAlloc = 200.000000
}

// ExampleMetGen_CollectCounterToChan демонстрирует, как воспользоваться методом CollectCounterToChan 
// для получения метрик типа Counter через канал.
func ExampleMetGen_CollectCounterToChan() {
	// Создаём контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаём экземпляр MetGen
	mg := metgen.New()

	// Добавим несколько тестовых метрик-счётчиков
	mg.MetricsCounter["PollCount"] = 10
	mg.MetricsCounter["RequestCount"] = 5

	// Создаём каналы для вывода метрик и ошибок
	outputChan := make(chan metgen.OneMetric)
	errChan := make(chan error)

	// Вызываем CollectCounterToChan в отдельной горутине
	go mg.CollectCounterToChan(ctx, outputChan, errChan)

	// Читаем данные из канала
	for metric := range outputChan {
		fmt.Printf("Counter Metric: %s = %f\n", metric.Name, metric.Metric)
	}

	// Читаем возможные ошибки (если они случились)
	select {
	case err := <-errChan:
		fmt.Printf("Error: %v\n", err)
	default:
		// Ошибок нет
	}

	// Output:
	// Counter Metric: RequestCount = 5.000000
	// Counter Metric: PollCount = 10.000000
}