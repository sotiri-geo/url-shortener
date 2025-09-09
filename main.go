package main

import (
	"log"
	"net/http"

	"github.com/sotiri-geo/url-shortener/internal/handler"
)

// hard coded implementation of store for now

type InMemoryURLStore struct {
	urls map[string]string
}

func (i *InMemoryURLStore) GetShortURL(url string) string {
	return "abc123"
}

func (i *InMemoryURLStore) GetOriginalURL(shortCode string) (string, bool) {
	url, exists := i.urls[shortCode]
	return url, exists
}

func main() {
	store := InMemoryURLStore{}
	shortener := handler.NewShortener(&store)
	// Create handle func and register route
	http.HandleFunc("/health", handler.HealthCheck)
	http.Handle("/shortener", shortener)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
