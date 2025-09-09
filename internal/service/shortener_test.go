package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/service"
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

	service.HealthCheck(response, req)

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
		server := service.NewShortener(&store)
		response := httptest.NewRecorder()
		want := "abc123"
		server.ServeHTTP(response, req)
		var got service.URLShortResponse

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
		server := service.NewShortener(&store)
		body := `{ invalid json }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := service.ErrorResponse{
			Error:   service.ERR_INVALID_JSON,
			Code:    service.ERR_INVALID_JSON_CODE,
			Details: service.ERR_INVALID_JSON_DETAILS,
		}
		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)

		// Check error response
		var got service.ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		assertErrorResponse(t, got, want)
	})

	t.Run("bad client request with empty url", func(t *testing.T) {
		store := FakeStore{}
		server := service.NewShortener(&store)
		body := `{ "url": "" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := service.ErrorResponse{
			Error:   service.ERR_EMPTY_URL,
			Code:    service.ERR_EMPTY_URL_CODE,
			Details: service.ERR_EMPTY_URL_DETAILS,
		}

		server.ServeHTTP(response, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)

		var got service.ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

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

func assertErrorResponse(t testing.TB, got, want service.ErrorResponse) {
	t.Helper()

	if got.Error != want.Error {
		t.Errorf("got error %q, want %q", got.Error, "empty URL")
	}
}
