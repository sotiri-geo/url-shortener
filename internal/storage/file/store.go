package file

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// mapping from shortCode to URL
type ShortCodeToURL map[string]string

type FileStore struct {
	Database *os.File
}

func (f *FileStore) Exists(shortCode string) (bool, error) {
	loaded, err := f.loadData()
	_, exists := loaded[shortCode]
	return exists, err
}

func (f *FileStore) Save(shortCode, originalUrl string) error {
	f.Database.Seek(0, io.SeekStart)
	// load
	loaded, err := f.loadData()
	if err != nil {
		return fmt.Errorf("failed to load data: %v", err)
	}
	// mutate
	loaded[shortCode] = originalUrl
	return json.NewEncoder(f.Database).Encode(&loaded)
}

func (f *FileStore) loadData() (ShortCodeToURL, error) {
	var data ShortCodeToURL
	// seek to the start
	_, err := f.Database.Seek(0, io.SeekStart)
	if err != nil {
		return data, fmt.Errorf("failed to seek start: %v", err)
	}
	err = json.NewDecoder(f.Database).Decode(&data)
	if err != nil {
		return data, fmt.Errorf("failed to decode: %v", err)
	}
	return data, nil
}
