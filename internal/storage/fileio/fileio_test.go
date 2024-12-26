package fileio

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


type mockLogger struct{}

func (m *mockLogger) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return len(p), nil
}

func TestNew(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 0))
	t.Run("Создание файла в существующей директории", func(t *testing.T) {
		tmpDir := t.TempDir()
		fileName := "testfile"

		f, err := New(tmpDir, fileName)
		require.NoError(t, err)
		assert.NotNil(t, f)
		defer f.Close()

		// Проверяем, что файл действительно создан
		_, statErr := os.Stat(filepath.Join(tmpDir, fileName))
		assert.NoError(t, statErr)
	})

	t.Run("Создание файла в несуществующей директории", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "some", "subdir")
		fileName := "testfile2"

		f, err := New(nestedDir, fileName)
		require.NoError(t, err)
		assert.NotNil(t, f)
		defer f.Close()

		_, statErr := os.Stat(filepath.Join(nestedDir, fileName))
		assert.NoError(t, statErr)
	})

	t.Run("Ошибка при создании директории", func(t *testing.T) {
		tmpDir := t.TempDir()
		fileName := "file_instead_of_dir"
		filePath := filepath.Join(tmpDir, fileName)

		// Создадим файл вместо каталога
		ftrash, err := os.Create(filePath)
		require.NoError(t, err)

		// Теперь New должно вернуть ошибку
		f, newErr := New(filePath, "anyfile")
		assert.Error(t, newErr)
		assert.Nil(t, f)
		ftrash.Close()
	})
}

func TestFile_Write(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 0))
	t.Run("Успешная запись в файл", func(t *testing.T) {
		tmpDir := t.TempDir()
		f, err := New(tmpDir, "test_write")
		require.NoError(t, err)
		defer f.Close()

		data := &Data{
			ItemsGauge:   map[string]float64{"g1": 1.23},
			ItemsCounter: map[string][]float64{"c1": {10, 20}},
		}

		err = f.Write(data)
		require.NoError(t, err)

		// Проверим напрямую, что записалось в файл (гоб-декод).
		file, err := os.Open(filepath.Join(tmpDir, "test_write"))
		require.NoError(t, err)
		defer file.Close()

		var readData Data
		dec := gob.NewDecoder(file)
		err = dec.Decode(&readData)
		require.NoError(t, err)

		assert.Equal(t, data.ItemsGauge, readData.ItemsGauge)
		assert.Equal(t, data.ItemsCounter, readData.ItemsCounter)
	})

	t.Run("Запись не происходит, если f.file == nil", func(t *testing.T) {
		f := &File{mu: &sync.Mutex{}}
		err := f.Write(&Data{
			ItemsGauge: map[string]float64{"niltest": 123},
		})
		require.NoError(t, err) // нет ошибки, просто пропускаем запись
	})
}

func TestFile_Read(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 5))
	t.Run("Успешное чтение из файла", func(t *testing.T) {
		tmpDir := t.TempDir()
		f, err := New(tmpDir, "test_read")
		require.NoError(t, err)
		defer f.Close()

		data := &Data{
			ItemsGauge:   map[string]float64{"g2": 9.99},
			ItemsCounter: map[string][]float64{"c2": {100, 200}},
		}
		// Сначала запишем, затем прочитаем через Read()
		err = f.Write(data)
		require.NoError(t, err)

		gauges, counters, readErr := f.Read()
		require.NoError(t, readErr)
		assert.Equal(t, data.ItemsGauge, gauges)
		assert.Equal(t, data.ItemsCounter, counters)
	})

	t.Run("Чтение при f.file == nil возвращает пустые мапы", func(t *testing.T) {
		f := &File{mu: &sync.Mutex{}}
		gauges, counters, err := f.Read()
		require.NoError(t, err)
		assert.Empty(t, gauges)
		assert.Empty(t, counters)
	})
}

func TestFile_Close(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 0))
	t.Run("Успешное закрытие файла", func(t *testing.T) {
		tmpDir := t.TempDir()
		f, err := New(tmpDir, "test_close")
		require.NoError(t, err)

		closeErr := f.Close()
		require.NoError(t, closeErr)
	})

	t.Run("Close при f.file == nil возвращает nil-ошибку", func(t *testing.T) {
		f := &File{mu: &sync.Mutex{}}
		err := f.Close()
		require.NoError(t, err)
	})
}

func Test_openFileRetry(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 0))
	t.Run("Успешное открытие файла (сразу или с созданием)", func(t *testing.T) {
		tmpDir := t.TempDir()
		fullPath := filepath.Join(tmpDir, "retry_test")

		file, err := openFileRetry(fullPath, os.O_RDWR|os.O_CREATE, 0644)
		require.NoError(t, err)
		require.NotNil(t, file)
		file.Close()

		_, statErr := os.Stat(fullPath)
		assert.NoError(t, statErr)
	})

	t.Run("Не удаётся открыть файл из-за неверного пути", func(t *testing.T) {
		// Пример пути, который с высокой вероятностью вызовет ошибку
		invalidPath := string(os.PathSeparator) + "root_denied" + string(os.PathSeparator) + "file"

		file, err := openFileRetry(invalidPath, os.O_RDWR, 0644)
		assert.Error(t, err)
		assert.Nil(t, file)
	})
}

func Test_writeToFileRetry(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 0))
	t.Run("Успешная запись (без ошибок)", func(t *testing.T) {
		tmpDir := t.TempDir()
		f, err := New(tmpDir, "write_retry_test")
		require.NoError(t, err)
		defer f.Close()

		data := &Data{
			ItemsGauge:   map[string]float64{"test": 3.14},
			ItemsCounter: map[string][]float64{"count": {1, 2, 3}},
		}
		err = f.writeToFileRetry(data)
		require.NoError(t, err)

		// Проверим результат
		file, err := os.Open(filepath.Join(tmpDir, "write_retry_test"))
		require.NoError(t, err)
		defer file.Close()

		var check Data
		require.NoError(t, gob.NewDecoder(file).Decode(&check))

		assert.Equal(t, data.ItemsGauge, check.ItemsGauge)
		assert.Equal(t, data.ItemsCounter, check.ItemsCounter)
	})

	t.Run("Ошибка при Truncate (или Seek)", func(t *testing.T) {
		tmpDir := t.TempDir()
	
		// Создаем файл только для чтения
		filePath := filepath.Join(tmpDir, "readonly_file")
		roFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0644)
		require.NoError(t, err)
	
		t.Cleanup(func() { roFile.Close() })
	
		f := &File{
			mu:   &sync.Mutex{},
			file: roFile, 
		}
	
		err = f.writeToFileRetry(&Data{})
		assert.Error(t, err)
	})
}

func Test_readFromFileRetry(t *testing.T) {
	assert.NoError(t, logger.Init(&mockLogger{}, 0))
	t.Run("Чтение из непустого файла (валидный gob)", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "read_retry_test")

		// Сразу запишем туда валидные данные
		{
			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
			require.NoError(t, err)

			data := &Data{
				ItemsGauge:   map[string]float64{"g": 2.71},
				ItemsCounter: map[string][]float64{"c": {42, 43}},
			}
			require.NoError(t, gob.NewEncoder(f).Encode(data))
			f.Close()
		}

		// Создаём структуру File и читаем
		fobj, err := New(tmpDir, "read_retry_test")
		require.NoError(t, err)
		defer fobj.Close()

		var data Data
		data.ItemsGauge = make(map[string]float64)
		data.ItemsCounter = make(map[string][]float64)

		err = fobj.readFromFileRetry(&data)
		require.NoError(t, err)
		assert.Equal(t, 2.71, data.ItemsGauge["g"])
		assert.Equal(t, []float64{42, 43}, data.ItemsCounter["c"])
	})

	t.Run("Чтение из пустого файла возвращает без ошибки", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty_file")
		_, err := os.Create(filePath)
		require.NoError(t, err)

		fobj, err := New(tmpDir, "empty_file")
		require.NoError(t, err)
		defer fobj.Close()

		var data Data
		data.ItemsGauge = make(map[string]float64)
		data.ItemsCounter = make(map[string][]float64)

		err = fobj.readFromFileRetry(&data)
		require.NoError(t, err)
		assert.Empty(t, data.ItemsGauge)
		assert.Empty(t, data.ItemsCounter)
	})

	t.Run("Чтение из повреждённого файла (файл удаляется)", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "corrupted_file")

		// Запишем мусор
		require.NoError(t, os.WriteFile(filePath, []byte("bad data"), 0644))

		fobj, err := New(tmpDir, "corrupted_file")
		require.NoError(t, err)
		defer fobj.Close()

		var data Data
		err = fobj.readFromFileRetry(&data)
		// Код удаляет файл и возвращает nil для data.Items...
		// При этом сам метод возвращает nil (не ошибку), если успешно удалил.
		require.NoError(t, err)
		assert.Nil(t, data.ItemsGauge)
		assert.Nil(t, data.ItemsCounter)

		// Проверяем, что файл удалён
		_, statErr := os.Stat(filePath)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("Ошибка Stat после всех попыток (например, некорректный диск)", func(t *testing.T) {
		// Можно создать структуру File с file == nil и вызвать readFromFileRetry напрямую:
		fobj := &File{
			file: nil,
			mu:   &sync.Mutex{},
		}
		var data Data
		err := fobj.readFromFileRetry(&data)
		assert.Error(t, err)
	})
}
