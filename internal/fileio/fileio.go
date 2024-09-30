package fileio

import (
	"io"
	"os"
	"path/filepath"
)

type File struct {
	Path string
	Name string
}

func (f *File) Write (data []byte) error {
	path := filepath.Join(f.Path, f.Name)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func (f *File) Read() ([]byte, error) {
	path := filepath.Join(f.Path, f.Name)

	file, err := os.Open(path)
	if err != nil {
        return nil, err
    }
	defer file.Close()

	return io.ReadAll(file)
}