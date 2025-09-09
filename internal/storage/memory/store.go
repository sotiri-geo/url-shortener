package memory

type MemoryDB struct {
	// shortUrl -> OriginalUrl
	urls map[string]string
}

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

func (m *MemoryDB) Save(shortUrl, originalUrl string) {
	m.urls[shortUrl] = originalUrl
}
