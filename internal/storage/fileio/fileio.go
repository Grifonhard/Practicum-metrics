package fileio

import (
	"encoding/gob"
	"fmt"
	//"io"

	"os"
	"path/filepath"
	//"strings"
	"sync"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
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
	var err error

	fullpath := filepath.Join(path, filename)

	if path != "" {
		err1 := os.Mkdir(path, 0755)
			if err1 != nil {
				logger.Error(err1)
				return nil, err1
			}
		file, err = os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE, 0666)
		//-------------костыль для тестов START-----------------------
		/*if err != nil && strings.Contains(err.Error(), "no such file or directory") {
			err1 := os.Mkdir(path, 0755)
			if err1 != nil {
				fmt.Println(1)
				logger.Error(err1)
			}
			file, err1 = os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE, 0666)
			if err1 != nil {
				fmt.Println(2)
				logger.Error(err1)
			}
			filesource, err1 := os.OpenFile("/tmp/crutchJJAOBAAF/backup", os.O_RDWR|os.O_CREATE, 0666)
			if err1 == nil {
				logger.Error(err1)
				defer filesource.Close()
				_, err1 = io.Copy(file, filesource)
				if err1 != nil {
					fmt.Println(3)
					logger.Error(err1)
				}
			} else {
				fmt.Println(4)
				fmt.Println(err1)
			}
			//-------------костыль для тестов END-----------------------
		} else */if err != nil {
			logger.Error(err)
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

	err := f.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("fail truncate file: %w", err)
	}

	_, err = f.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to move pointer to beginning of file: %w", err)
	}

	encoder := gob.NewEncoder(f.file)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
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

	fileInfo, err := f.file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось получить информацию о файле: %w", err)
	}

	if fileInfo.Size() == 0 {
		return data.ItemsGauge, data.ItemsCounter, nil
	}

	decoder := gob.NewDecoder(f.file)

	err = decoder.Decode(&data)
	if err != nil {
		logger.Error(fmt.Sprintf("Файл %s поврежден, удаляю...\n", f.fullpath))
		removeErr := os.Remove(f.fullpath)
		if removeErr != nil {
			return nil, nil, fmt.Errorf("не удалось удалить поврежденный файл: %w", removeErr)
		}
		logger.Error(err)
		return data.ItemsGauge, data.ItemsCounter, nil
	}

	return data.ItemsGauge, data.ItemsCounter, nil
}

func (f *File) Close() error {
	if f.file == nil {
		return nil
	}
	//-------------костыль для тестов START-----------------------
	/*logger.Info("create directory /tmp/crutchJJAOBAAF")
	err := os.Mkdir("/tmp/crutchJJAOBAAF", 0755)
	if err != nil {
		logger.Error(err)
		return f.file.Close()
	}
	filenew, err := os.OpenFile("/tmp/crutchJJAOBAAF/backup", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		logger.Error(err)
		return f.file.Close()
	}
	defer filenew.Close()
	_, err = io.Copy(filenew, f.file)
	if err != nil {
		logger.Error(err)
		return f.file.Close()
	}*/
	//-------------костыль для тестов END-----------------------
	return f.file.Close()
}
