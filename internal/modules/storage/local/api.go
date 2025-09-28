package local

import (
	"io"
	"os"
	"path/filepath"
)

func SaveFile(f io.Reader, path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0770)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, f)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFile(path string) error {
	return os.Remove(path)
}
