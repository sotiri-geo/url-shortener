package storage

type URLStore interface {
	Exists(shortCode string) bool
	Save(shortCode, originalUrl string) error
	GetOriginalURL(shortUrl string) (string, bool)
}
