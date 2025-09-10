package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/generator"
	"github.com/sotiri-geo/url-shortener/internal/handler"
	"github.com/sotiri-geo/url-shortener/internal/storage/memory"
)

func TestGeneratingShortnerAndRedirector(t *testing.T) {
	cases := []struct {
		name            string
		originalUrl     string
		setupGen        func(g *generator.RandomChars)
		setupStore      func(s *memory.MemoryDB)
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "create short code and use it to be redirected to original url",
			originalUrl:     "https://google.com",
			setupGen:        func(g *generator.RandomChars) {},
			setupStore:      func(s *memory.MemoryDB) {},
			wantStatus:      http.StatusFound,
			wantContentType: handler.JsonContentType,
		},
	}

	for _, tt := range cases {
		// Setup
		store := memory.New()
		gen := generator.New(generator.RandomGenSize)
		shortner := handler.NewShortener(store, gen)
		redirector := handler.NewRedirector(store)

		// execute - shorten server
		request := newShortenRequest(fmt.Sprintf(`{"url": "%s"}`, tt.originalUrl))
		response := httptest.NewRecorder()
		shortner.ServeHTTP(response, request)

		// assert - shorten behaviour
		assertStatusCode(t, response.Code, http.StatusCreated)
		assertContentType(t, response.Result().Header.Get("content-type"), handler.JsonContentType)

		shortCode, err := getShortCode(response.Body)
		assertNoErr(t, err)

		// execute - redirect
		request = newRedirectRequest(shortCode.Short)
		response = httptest.NewRecorder()
		redirector.ServeHTTP(response, request)

		// assert - redirect behaviour
		assertStatusCode(t, response.Code, http.StatusFound)
		assertContentType(t, response.Result().Header.Get("content-type"), "text/html; charset=utf-8")
		assertLocationHeader(t, response.Header().Get("Location"), "https://google.com")
	}
}
