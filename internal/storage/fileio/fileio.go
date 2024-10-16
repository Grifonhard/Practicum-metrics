package fileio

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	"github.com/bytedance/sonic/decoder"
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

// если неудачно
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

func New(path, filename string) (*File, error) {
	var mu sync.Mutex
	var file *os.File
	var err error
	var errCollect []error

	fullpath := filepath.Join(path, filename)
	if path != "" {
		err = os.Mkdir(path, 0755)
		if err != nil && !os.IsExist(err){
			return nil, err
		}
		for i := 0; i < MAXRETRIES; i++ {
			file, err = os.OpenFile(fullpath, os.O_RDWR, 0666)
			if os.IsNotExist(err) {
				file, err = os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE, 0666)
			}
			if err != nil{
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				errCollect = append(errCollect, err)
				continue
			} else {
				break
			}
		}
		if errCollect != nil {
			fmt.Printf("problem with open file: %s\n", errors.Join(errCollect...).Error())
		}
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
	var errCollect []error
	var err error

	for i := 0; i < MAXRETRIES; i++ {
		err = f.file.Truncate(0)
		if err != nil{
			err = fmt.Errorf("fail truncate file: %w", err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			errCollect = append(errCollect, err)
			continue
		}
		
		_, err = f.file.Seek(0, 0)
		if err != nil{
			err = fmt.Errorf("failed to move pointer to beginning of file: %w", err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			errCollect = append(errCollect, err)
			continue
		}

		encoder := gob.NewEncoder(f.file)

		if err := encoder.Encode(data); err != nil {
			err = fmt.Errorf("failed to write data to file: %w", err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			errCollect = append(errCollect, err)
			continue
		} else {
			break
		}
	}
	if errCollect != nil {
		fmt.Printf("problem with write file: %s\n", errors.Join(errCollect...).Error())
	}
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
	var errCollect []error
	var err error

	for i := 0; i < MAXRETRIES + 1; i++ {
		fileInfo, err := f.file.Stat()
		if err != nil {
			if i == MAXRETRIES {
				err = fmt.Errorf("не удалось получить информацию о файле: %w", err)
				break
			}
			err = fmt.Errorf("не удалось получить информацию о файле: %w", err)
			errCollect = append(errCollect, err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			continue
		}

		if fileInfo.Size() == 0 {
			return data.ItemsGauge, data.ItemsCounter, nil
		}

		decoder := gob.NewDecoder(f.file)

		err = decoder.Decode(&data)
		if err != nil {
			if i == MAXRETRIES {
				fmt.Printf("problem with read file: %s\n", errors.Join(errCollect...).Error())
				logger.Error(fmt.Sprintf("Файл %s поврежден, удаляю...\n", f.fullpath))
				removeErr := os.Remove(f.fullpath)
				if removeErr != nil {
					return nil, nil, fmt.Errorf("не удалось удалить поврежденный файл: %w", removeErr)
				}
				logger.Error(err)
				return data.ItemsGauge, data.ItemsCounter, nil
			}
			err = fmt.Errorf("проблемы с чтением файла: %w", err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			continue
		} else {
			break
		}
	}
	if errCollect != nil {
		fmt.Printf("problem with read file: %s\n", errors.Join(errCollect...).Error())
	}
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
