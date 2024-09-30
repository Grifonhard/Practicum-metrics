package fileio

import (
	"encoding/gob"
	"os"
	"path/filepath"
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
	path := filepath.Join(f.Path, f.Name)

	file, err := os.Open(path)
	if err != nil {
        return nil, nil, err
    }
	defer file.Close()

	decoder := gob.NewDecoder(file)

	var data Data
	data.ItemsGauge = make(map[string]float64)
	data.ItemsCounter = make(map[string][]float64)

	err = decoder.Decode(&data)
	if err != nil {
        return nil, nil, err
    }
	
	return data.ItemsGauge, data.ItemsCounter, nil
}