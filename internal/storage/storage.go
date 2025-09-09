package storage

type URLStore interface {
	Save(shortCode, original string)
	GetShortURL(url string) string
	GetOriginalURL(shortCode string) (string, bool)
}
