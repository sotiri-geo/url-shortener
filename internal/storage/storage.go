package storage

type URLStore interface {
	Save(shortCode, originalUrl string)
	GetOriginalURL(shortUrl string) (string, bool)
}
