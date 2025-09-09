package memory

import "errors"

type MemoryDB struct {
	// shortUrl -> OriginalUrl
	urls map[string]string
}

var ErrShortUrlExists = errors.New("short url already exists in store")

func New() *MemoryDB {
	return &MemoryDB{urls: make(map[string]string)}
}

func NewWithData(urls map[string]string) *MemoryDB {
	return &MemoryDB{urls}
}

func (m *MemoryDB) GetOriginalUrl(shortUrl string) (string, bool) {
	original, exists := m.urls[shortUrl]
	return original, exists
}

func (m *MemoryDB) Save(shortUrl, originalUrl string) error {
	_, exists := m.GetOriginalUrl(shortUrl)
	if exists {
		return ErrShortUrlExists
	}
	m.urls[shortUrl] = originalUrl
	return nil
}
