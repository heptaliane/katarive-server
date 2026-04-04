package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

func url2filename(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
func NewFile(path string) (*os.File, error) {
	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return nil, err
	}

	return os.Create(path)
}

func LoadJson[T any](path string) (*T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := json.NewDecoder(f)

	var data T
	if err := d.Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
func DumpJson[T any](path string, data *T) error {
	f, err := NewFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	e := json.NewEncoder(f)
	return e.Encode(data)
}
