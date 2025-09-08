package main

import (
	"log"
	"net/http"
)

// hard coded implementation of store for now

type InMemoryURLStore struct{}

func (i *InMemoryURLStore) GetShortURL(url string) string {
	return "abc123"
}

func main() {

	// Create handle func and register route
	http.HandleFunc("/health", HealthCheck)
	http.Handle("/shortener", &URLServer{&InMemoryURLStore{}})
	log.Fatal(http.ListenAndServe(":3000", nil))
}
