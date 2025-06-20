package tools

import "os"

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func PanicOnError[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
