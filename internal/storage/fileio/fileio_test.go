package fileio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "testdata.gob")

	// Создаем файл и структуру данных
	file := New(tempDir, "testdata.gob")
	dataToWrite := &Data{
		ItemsGauge: map[string]float64{
			"gauge1": 123.45,
		},
		ItemsCounter: map[string][]float64{
			"counter1": {1.1, 2.2, 3.3},
		},
	}

	// Тестируем запись данных в файл
	err := file.Write(dataToWrite)
	require.NoError(t, err, "ошибка при записи данных в файл")

	// Проверяем, что файл был создан
	_, err = os.Stat(filePath)
	assert.NoError(t, err, "файл не был создан")
}

func TestRead(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем файл и структуру данных
	file := New(tempDir, "testdata.gob")
	dataToWrite := &Data{
		ItemsGauge: map[string]float64{
			"gauge1": 123.45,
		},
		ItemsCounter: map[string][]float64{
			"counter1": {1.1, 2.2, 3.3},
		},
	}

	// Записываем данные для последующего чтения
	err := file.Write(dataToWrite)
	require.NoError(t, err, "ошибка при записи данных в файл")

	// Тестируем чтение данных из файла
	itemsGauge, itemsCounter, err := file.Read()
	require.NoError(t, err, "ошибка при чтении данных из файла")

	// Проверяем, что прочитанные данные совпадают с записанными
	assert.Equal(t, dataToWrite.ItemsGauge, itemsGauge, "не совпадают значения gauge")
	assert.Equal(t, dataToWrite.ItemsCounter, itemsCounter, "не совпадают значения counter")
}