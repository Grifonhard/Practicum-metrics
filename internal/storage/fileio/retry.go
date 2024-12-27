package fileio

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
)

// Настройки повторных действий в случае неудачных попыток чтения/записи
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

// openFileRetry открывает файл для чтения/записи
// используется в New
// повторяет если не удалось
func openFileRetry(name string, flag int, perm fs.FileMode) (file *os.File, err error) {
	var errCollect []error
	for i := 0; i < MAXRETRIES; i++ {
		file, err = os.OpenFile(name, flag, perm)
		if os.IsNotExist(err) {
			file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm)
		}
		if err != nil {
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			errCollect = append(errCollect, err)
			continue
		} else {
			break
		}
	}
	if errCollect != nil {
		logger.Error(fmt.Sprintf("problem with open file: %s\n", errors.Join(errCollect...).Error()))
	}
	return file, err
}

// writeToFileRetry запись в файл
// используется в Write
// повторяет если не удалось
func (f *File) writeToFileRetry(data *Data) error {
	var errCollect []error
	var err error
	if f.file == nil {
		return ErrFileNil
	}
	for i := 0; i < MAXRETRIES; i++ {
		err = f.file.Truncate(0)
		if err != nil {
			err = fmt.Errorf("fail truncate file: %w", err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			errCollect = append(errCollect, err)
			continue
		}

		_, err = f.file.Seek(0, 0)
		if err != nil {
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
		logger.Error(fmt.Sprintf("problem with write file: %s\n", errors.Join(errCollect...).Error()))
	}
	return err
}

// readFromFileRetry чтение из файла
// используется в Read
// повторение если не удалось
func (f *File) readFromFileRetry(data *Data) (err error) {
	var errCollect []error
	var fileInfo fs.FileInfo
	if f.file == nil {
		return ErrFileNil
	}
	for i := 0; i < MAXRETRIES+1; i++ {
		fileInfo, err = f.file.Stat()
		if err != nil {
			if i == MAXRETRIES {
				err = fmt.Errorf("не удалось получить информацию о файле: %w", err)
				errCollect = append(errCollect, err)
				break
			}
			err = fmt.Errorf("не удалось получить информацию о файле: %w", err)
			errCollect = append(errCollect, err)
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			continue
		}

		if fileInfo.Size() == 0 {
			return nil
		}

		_, err = f.file.Seek(0, 0)
		if err != nil {
			err = fmt.Errorf("не удалось переместить курсор в начало файла: %w", err)
			errCollect = append(errCollect, err)
			if i == MAXRETRIES {
				break
			}
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			continue
		}

		decoder := gob.NewDecoder(f.file)

		err = decoder.Decode(&data)
		if err != nil && !errors.Is(err, io.EOF) {
			if i == MAXRETRIES {
				logger.Error(fmt.Sprintf("problem with read file: %s\n", errors.Join(errCollect...).Error()))
				logger.Error(fmt.Sprintf("Файл %s поврежден, удаляю...\n", f.fullpath))
				removeErr := os.Remove(f.fullpath)
				if removeErr != nil {
					data.ItemsCounter = nil
					data.ItemsGauge = nil
					return fmt.Errorf("не удалось удалить поврежденный файл: %w", removeErr)
				}
				data.ItemsCounter = nil
				data.ItemsGauge = nil
				logger.Error(err)
				return nil
			}
			err = fmt.Errorf("проблемы с чтением файла: %w", err)
			errCollect = append(errCollect, err)
			data.ItemsCounter = nil
			data.ItemsGauge = nil
			time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
			continue
		} else {
			break
		}
	}
	if errCollect != nil {
		logger.Error(fmt.Sprintf("problem with read file: %s\n", errors.Join(errCollect...).Error()))
	}
	return nil
}
