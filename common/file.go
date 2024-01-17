package common

import (
	"io"
	"os"
)

// SaveTmpFile saves data to a temporary file. The filename would be apppended with a random string.
func SaveTmpFile(filename string, data io.Reader) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, data)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}
