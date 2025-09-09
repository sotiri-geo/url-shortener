package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/handler"
)

func TestGeneratingShortCodesAndRedirecting(t *testing.T) {

	store := NewFakeStore()
	shortner := handler.NewShortener(store, NewStubGenerator())
	redirector := handler.NewRedirector(store)

	t.Run("store url then use short url to be redirected", func(t *testing.T) {
		response := httptest.NewRecorder()
		shortner.ServeHTTP(response, newShortenRequest(`{"url": "https://google.com"}`))

		assertStatusCode(t, response.Code, http.StatusCreated)
		assertContentType(t, response.Header().Get("content-type"), handler.JsonContentType)
		URL, err := getShortCode(response.Body)

		if err != nil {
			t.Fatalf("failed to parse response from /shorten: %v", err)
		}

		// attempt redirect
		response = httptest.NewRecorder()
		req := newRedirectRequest(URL.Short)
		redirector.ServeHTTP(response, req)

		assertStatusCode(t, response.Code, http.StatusFound)
		assertContentType(t, response.Result().Header.Get("content-type"), "text/html; charset=utf-8")
		assertLocationHeader(t, response.Header().Get("Location"), "https://google.com")

	})
}
