package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestHealthCheckEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	HealthCheck(response, req)

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
		store := FakeStore{}
		server := NewShortener(&store)
		response := httptest.NewRecorder()
		want := "abc123"
		server.ServeHTTP(response, req)
		var got URLShortResponse

		// decode
		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		assertStatusCode(t, response.Code, http.StatusCreated)
		assertURL(t, got.Short, want)
	})

	t.Run("bad client request with missing url key", func(t *testing.T) {
		store := FakeStore{}
		server := NewShortener(&store)
		body := `{ invalid json }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := ErrorResponse{
			Error:   ERR_INVALID_JSON,
			Code:    ERR_INVALID_JSON_CODE,
			Details: ERR_INVALID_JSON_DETAILS,
		}
		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)

		// Check error response
		var got ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		assertErrorResponse(t, got, want)
	})

	t.Run("bad client request with empty url", func(t *testing.T) {
		store := FakeStore{}
		server := NewShortener(&store)
		body := `{ "url": "" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := ErrorResponse{
			Error:   ERR_EMPTY_URL,
			Code:    ERR_EMPTY_URL_CODE,
			Details: ERR_EMPTY_URL_DETAILS,
		}

		server.ServeHTTP(response, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)

		var got ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

		assertErrorResponse(t, got, want)
	})

	t.Run("GET /abc123 redirects client to location", func(t *testing.T) {
		store := FakeStore{urls: map[string]string{"abc123": "https://example.com"}}
		server := NewShortener(&store)
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
		server := NewShortener(&store)
		req := httptest.NewRequest(http.MethodGet, "/xyz123", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusNotFound)

		want := ErrorResponse{Error: "short code not found"}

		var got ErrorResponse

		json.NewDecoder(response.Body).Decode(&got)
		assertErrorResponse(t, got, want)

	})
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

func assertErrorResponse(t testing.TB, got, want ErrorResponse) {
	t.Helper()

	if got.Error != want.Error {
		t.Errorf("got error %q, want %q", got.Error, "empty URL")
	}
}
