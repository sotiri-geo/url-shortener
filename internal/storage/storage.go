package storage

type URLStore interface {
	Save(shortUrl, originalUrl string)
	GetOriginalURL(shortUrl string) (string, bool)
}
