package memory

import "errors"

type MemoryDB struct {
	// shortCode -> OriginalUrl
	urls map[string]string
}

var ErrShortCodeExists = errors.New("short code already exists in store")

func New() *MemoryDB {
	return &MemoryDB{urls: make(map[string]string)}
}

func NewWithData(urls map[string]string) *MemoryDB {
	return &MemoryDB{urls}
}

func (m *MemoryDB) GetOriginalURL(shortCode string) (string, bool) {
	original, exists := m.urls[shortCode]
	return original, exists
}

func (m *MemoryDB) Save(shortCode, originalUrl string) error {
	_, exists := m.GetOriginalURL(shortCode)
	if exists {
		return ErrShortCodeExists
	}
	m.urls[shortCode] = originalUrl
	return nil
}

func (m *MemoryDB) Exists(shortCode string) bool {
	_, exists := m.urls[shortCode]
	return exists
}
