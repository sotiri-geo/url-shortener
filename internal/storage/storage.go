package storage

type URLStore interface {
	Exists(shortCode string) bool
	Save(shortCode, originalUrl string)
	GetOriginalURL(shortUrl string) (string, bool)
}
