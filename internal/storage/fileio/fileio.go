package fileio

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
)

type File struct {
	Path string
	Name string
}

type Data struct {
	ItemsGauge   map[string]float64
	ItemsCounter map[string][]float64
}

func New(path, filename string) *File {
	return &File{
		Path: path,
		Name: filename,
	}
}

func (f *File) Write(data *Data) error {
	path := filepath.Join(f.Path, f.Name)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)

	return encoder.Encode(data)
}

func (f *File) Read() (map[string]float64, map[string][]float64, error) {
	var data Data
	data.ItemsGauge = make(map[string]float64)
	data.ItemsCounter = make(map[string][]float64)

	path := filepath.Join(f.Path, f.Name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return data.ItemsGauge, data.ItemsCounter, nil
	} else if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(path)
	if err != nil {
        return nil, nil, err
    }
	defer file.Close()

	decoder := gob.NewDecoder(file)

	err = decoder.Decode(&data)
	if err != nil {
		logger.Error(fmt.Sprintf("Файл %s поврежден, удаляю...\n", path))
		removeErr := os.Remove(path)
		if removeErr != nil {
			return nil, nil, fmt.Errorf("не удалось удалить поврежденный файл: %w", removeErr)
		}
	}
	
	return data.ItemsGauge, data.ItemsCounter, nil
}