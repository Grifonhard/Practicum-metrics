package fileio

import (
	"os"
	"path/filepath"
	"sync"
)

type File struct {
	fullpath string
	file     *os.File
	mu       *sync.Mutex
}

type Data struct {
	ItemsGauge   map[string]float64
	ItemsCounter map[string][]float64
}

func New(path, filename string) (*File, error) {
	var mu sync.Mutex
	var file *os.File

	fullpath := filepath.Join(path, filename)
	if path != "" {
		err := os.Mkdir(path, 0755)
		if err != nil && !os.IsExist(err){
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

func (f *File) Close() error {
	if f.file == nil {
		return nil
	}
	return f.file.Close()
}
