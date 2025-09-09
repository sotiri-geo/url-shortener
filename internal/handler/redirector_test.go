package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/handler"
)

func TestRedirector(t *testing.T) {
	t.Run("GET /abc123 redirects client to location", func(t *testing.T) {
		store := FakeStore{urls: map[string]string{"abc123": "https://example.com"}}
		server := handler.NewRedirector(&store)
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusFound)

		// Test location headers for redirect
		got := response.Header().Get("Location")
		want := "https://example.com"

		if got != want {
			t.Errorf("got location %q, want %q", got, want)
		}

	})

	t.Run("GET /xyz123 redirect not found location", func(t *testing.T) {
		store := FakeStore{urls: map[string]string{"abc123": "https://example.com"}}
		server := handler.NewRedirector(&store)
		req := httptest.NewRequest(http.MethodGet, "/xyz123", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusNotFound)

		want := handler.ErrorResponse{Error: "short code not found"}

		var got handler.ErrorResponse

		json.NewDecoder(response.Body).Decode(&got)
		assertErrorResponse(t, got, want)

	})

}
