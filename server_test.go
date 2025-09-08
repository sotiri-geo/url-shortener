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

func TestShortenURL(t *testing.T) {
	t.Run("POST /shorten returns a shortened url", func(t *testing.T) {
		body := `{ "url": "https://example.com" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))

		server := URLServer{&FakeStore{}}
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
		assertShortenURL(t, got.Short, want)
	})

	t.Run("bad client request with missing url key", func(t *testing.T) {
		server := URLServer{&FakeStore{}}
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
		server := URLServer{&FakeStore{}}
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

	t.Run("GET /abc123 redirects client", func(t *testing.T) {
		server := URLServer{&FakeStore{urls: map[string]string{"abc123": "https://example.com"}}}
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatusCode(t, response.Code, http.StatusFound)

	})
}

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("incorrect status code: got %d, want %d", got, want)
	}
}

func assertShortenURL(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("not the same shortened url: got %q, want %q", got, want)
	}
}

func assertErrorResponse(t testing.TB, got, want ErrorResponse) {
	t.Helper()

	if got.Error != want.Error {
		t.Errorf("got error %q, want %q", got.Error, "empty URL")
	}

	if got.Code != want.Code {
		t.Errorf("got error code %q, want %q", got.Code, "EMPTY_URL")
	}

	if got.Details != want.Details {
		t.Errorf("got error details %q, want %q", got.Details, "url must not be empty")
	}
}
