package fileio

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	//"github.com/Grifonhard/Practicum-metrics/internal/logger"
)

type File struct {
	fullpath string
	file     io.ReadWriter
	mu       *sync.Mutex
}

type Data struct {
	ItemsGauge   map[string]float64
	ItemsCounter map[string][]float64
}

func New(path, filename string) (*File, error) {
	var mu sync.Mutex
	fullpath := filepath.Join(path, filename)

	_, err := os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE, 0666)
	if strings.Contains(err.Error(), "no such file or directory") {
		err = os.Mkdir(path, 0755)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}

	/*file, err := os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}*/
	var buf bytes.Buffer

	fmt.Println(fullpath)

	return &File{
		fullpath: fullpath,
		file:     &buf,
		mu:       &mu,
	}, nil
}

func (f *File) Write(data *Data) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	/*err := f.file.Truncate(0)
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

	return f.file.Sync()*/
	return nil
}

func (f *File) Read() (map[string]float64, map[string][]float64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var data Data
	data.ItemsGauge = make(map[string]float64)
	data.ItemsCounter = make(map[string][]float64)

	/*fileInfo, err := f.file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось получить информацию о файле: %w", err)
	}

	if fileInfo.Size() == 0 {
		return data.ItemsGauge, data.ItemsCounter, nil
	}*/

	decoder := gob.NewDecoder(f.file)

	err := decoder.Decode(&data)
	if err != nil {
		/*logger.Error(fmt.Sprintf("Файл %s поврежден, удаляю...\n", f.fullpath))
		removeErr := os.Remove(f.fullpath)
		if removeErr != nil {
			return nil, nil, fmt.Errorf("не удалось удалить поврежденный файл: %w", removeErr)
		}*/
		fmt.Println(err)
	}

	return data.ItemsGauge, data.ItemsCounter, nil
}

func (f *File) Close() error {
	//return f.file.Close()
	return nil
}
