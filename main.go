package main

import (
	"log"
	"net/http"

	"github.com/sotiri-geo/url-shortener/internal/generator"
	"github.com/sotiri-geo/url-shortener/internal/handler"
	"github.com/sotiri-geo/url-shortener/internal/storage/memory"
)

func main() {
	store := memory.New()
	gen := generator.New(generator.RandomGenSize)
	shortener := handler.NewShortener(store, gen)
	// Create handle func and register route
	http.HandleFunc("/health", handler.HealthCheck)
	http.Handle("/shortener", shortener)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
