package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/handler"
	"github.com/sotiri-geo/url-shortener/internal/storage"
)

type FakeStore struct {
	urls map[string]string
}

func (f *FakeStore) GetShortURL(url string) string {
	return "abc123"
}

func (f *FakeStore) GetOriginalURL(shortCode string) (string, bool) {
	url, exists := f.urls[shortCode]
	return url, exists
}

func (f *FakeStore) Save(shortCode, original string) {
	f.urls[shortCode] = original
}

func NewFakeStore() *FakeStore {
	return &FakeStore{urls: make(map[string]string)}
}

func TestHealthCheckEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handler.HealthCheck(response, req)

	got := response.Body.String()

	want := "ok"

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusOK)
	}

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestURL(t *testing.T) {
	t.Run("POST /shorten returns a shortened url", func(t *testing.T) {
		body := `{ "url": "https://example.com" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		store := NewFakeStore()
		server := handler.NewShortener(store)
		response := httptest.NewRecorder()
		want := "abc123"
		server.ServeHTTP(response, req)
		var got handler.URLShortResponse

		// decode
		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		assertStatusCode(t, response.Code, http.StatusCreated)
		assertContentType(t, response.Result().Header.Get("content-type"), "application/json")
		assertURL(t, got.Short, want)
	})

	t.Run("POST /shorten stores state", func(t *testing.T) {
		body := `{ "url": "https://example.com" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		store := NewFakeStore()
		server := handler.NewShortener(store)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		var got handler.URLShortResponse

		// decode
		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if len(store.urls) != 1 {
			t.Errorf("did not store short URL: got length %d, want %d", len(store.urls), 1)
		}
		assertShortCodeStored(t, store, got.Short, "https://example.com")
	})

	t.Run("bad client request with missing url key", func(t *testing.T) {
		store := NewFakeStore()
		server := handler.NewShortener(store)
		body := `{ invalid json }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := handler.ErrorResponse{
			Error:   handler.ERR_INVALID_JSON,
			Code:    handler.ERR_INVALID_JSON_CODE,
			Details: handler.ERR_INVALID_JSON_DETAILS,
		}
		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)

		// Check error response
		var got handler.ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		assertErrorResponse(t, got, want)
	})

	t.Run("bad client request with empty url", func(t *testing.T) {
		store := NewFakeStore()
		server := handler.NewShortener(store)
		body := `{ "url": "" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := handler.ErrorResponse{
			Error:   handler.ERR_EMPTY_URL,
			Code:    handler.ERR_EMPTY_URL_CODE,
			Details: handler.ERR_EMPTY_URL_DETAILS,
		}

		server.ServeHTTP(response, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)

		var got handler.ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

		assertErrorResponse(t, got, want)
	})

}

func assertShortCodeStored(t testing.TB, store storage.URLStore, shortCode, wantOriginal string) {
	t.Helper()
	gotOriginal, exists := store.GetOriginalURL(shortCode)

	if !exists {
		t.Errorf("could not find short code: %q", shortCode)
	}

	if gotOriginal != "https://example.com" {
		t.Errorf("incorrect state stored: got %q, want %q", gotOriginal, wantOriginal)
	}

}

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("incorrect status code: got %d, want %d", got, want)
	}
}

func assertURL(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("not the same url: got %q, want %q", got, want)
	}
}

func assertErrorResponse(t testing.TB, got, want handler.ErrorResponse) {
	t.Helper()

	if got.Error != want.Error {
		t.Errorf("got error %q, want %q", got.Error, "empty URL")
	}
}

func assertContentType(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("incorrect content type: got %q, want %q", got, want)
	}
}
