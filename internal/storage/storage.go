package storage

type URLStore interface {
	GetShortURL(url string) string
	GetOriginalURL(shortCode string) (string, bool)
}
