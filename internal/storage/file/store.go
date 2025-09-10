package file

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/sotiri-geo/url-shortener/internal/handler"
)

type FileStore struct {
	Database io.ReadWriteSeeker
	urls     handler.URL
}

func (f *FileStore) Exists(shortCode string) (bool, error) {
	err := f.loadFromCache()
	if err != nil {
		return false, fmt.Errorf("failed to execute method Exists(%q): %v", shortCode, err)
	}
	_, exists := f.urls[shortCode]
	return exists, nil
}

func (f *FileStore) GetOriginalURL(shortCode string) (string, error) {
	err := f.loadFromCache()
	if err != nil {
		return "", fmt.Errorf("failed to get original url from short code %q: %v", shortCode, err)
	}

	return f.urls[shortCode], nil
}

func (f *FileStore) Save(shortCode, originalUrl string) error {
	urls, err := f.loadFromDisk()
	if err != nil {
		return fmt.Errorf("failed to loadFromDisk before saving: %v", err)
	}

	urls[shortCode] = originalUrl

	// seek to the beginning and rewrite
	f.Database.Seek(0, io.SeekStart)
	err = json.NewEncoder(f.Database).Encode(&urls)

	if err != nil {
		return fmt.Errorf("failed to encode: %v", err)
	}
	return err
}

func (f *FileStore) loadFromDisk() (handler.URL, error) {
	f.Database.Seek(0, io.SeekStart)
	urls, err := handler.NewUrls(f.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to read urls: %v", err)
	}
	return urls, err
}

// Behaves like a read through cache
func (f *FileStore) loadFromCache() error {
	if f.urls == nil {
		// force load from disk
		urls, err := f.loadFromDisk()
		if err != nil {
			return err
		}
		f.urls = urls
	}
	return nil
}

func NewFileStore(database io.ReadWriteSeeker) *FileStore {
	return &FileStore{Database: database}
}
