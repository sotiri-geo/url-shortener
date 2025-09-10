package file

import (
	"fmt"
	"io"

	"github.com/sotiri-geo/url-shortener/internal/handler"
)

type FileStore struct {
	Database io.ReadWriteSeeker
}

func (f *FileStore) Exists(shortCode string) (bool, error) {
	urls, err := f.load()
	if err != nil {
		return false, fmt.Errorf("failed to execute method Exists(%q): %v", shortCode, err)
	}
	_, exists := urls[shortCode]
	return exists, nil
}

func (f *FileStore) GetOriginalURL(shortCode string) (string, error) {
	urls, err := f.load()
	if err != nil {
		return "", fmt.Errorf("failed to get original url from short code %q: %v", shortCode, err)
	}

	return urls[shortCode], nil
}

func (f *FileStore) load() (handler.URL, error) {
	f.Database.Seek(0, io.SeekStart)
	urls, err := handler.NewUrls(f.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to read urls: %v", err)
	}
	return urls, err
}

func NewFileStore(database io.ReadWriteSeeker) *FileStore {
	return &FileStore{Database: database}
}
