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
		req := newRedirectRequest("abc123")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusFound)
		assertContentType(t, response.Result().Header.Get("content-type"), "text/html; charset=utf-8")
		assertLocationHeader(t, response.Header().Get("Location"), "https://example.com")
	})

	t.Run("GET /xyz123 redirect not found location", func(t *testing.T) {
		store := FakeStore{urls: map[string]string{"abc123": "https://example.com"}}
		server := handler.NewRedirector(&store)
		req := newRedirectRequest("xyz123")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusNotFound)

		want := handler.ErrorResponse{Error: "short code not found"}

		var got handler.ErrorResponse

		json.NewDecoder(response.Body).Decode(&got)
		assertErrorResponse(t, got, want)

	})

}

func newRedirectRequest(shortCode string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
}

func assertLocationHeader(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got location %q, want %q", got, want)
	}
}
