// Модуль отвечает за сохранение/чтение бэкапа данных о метриках
package fileio

import (
	"os"
	"path/filepath"
	"sync"
)

// File хранит данные о файле для бэкапов
type File struct {
	fullpath string
	file     *os.File
	mu       *sync.Mutex
}

// Data в этом формате данные передаются выше
type Data struct {
	ItemsGauge   map[string]float64
	ItemsCounter map[string][]float64
}

// New создание нового экземпляра file
func New(path, filename string) (*File, error) {
	var mu sync.Mutex
	var file *os.File

	fullpath := filepath.Join(path, filename)
	if path != "" {
		err := os.MkdirAll(path, 0755)
		if err != nil && !os.IsExist(err) {
			return nil, err
		}
		file, err = openFileRetry(fullpath, os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
	}

	return &File{
		fullpath: fullpath,
		file:     file,
		mu:       &mu,
	}, nil
}

// Write запись данных в файл
// данные кодируются в gob
func (f *File) Write(data *Data) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.file == nil {
		return nil
	}

	err := f.writeToFileRetry(data)
	if err != nil {
		return err
	}

	return f.file.Sync()
}

// Read чтение данных из файла
func (f *File) Read() (map[string]float64, map[string][]float64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var data Data
	data.ItemsGauge = make(map[string]float64)
	data.ItemsCounter = make(map[string][]float64)

	if f.file == nil {
		return data.ItemsGauge, data.ItemsCounter, nil
	}

	err := f.readFromFileRetry(&data)
	if err != nil {
		return nil, nil, err
	}

	return data.ItemsGauge, data.ItemsCounter, nil
}

// Close закрытие файла, открытого в New
func (f *File) Close() error {
	if f.file == nil {
		return nil
	}
	return f.file.Close()
}
