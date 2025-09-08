package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// type HandlerFunc func(ResponseWriter, *Request)

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

		response := httptest.NewRecorder()
		want := "abc123"
		URLServer(response, req)
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
		body := `{ invalid json }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := ErrorResponse{
			Error:   "invalid JSON",
			Code:    "INVALID_JSON",
			Details: "not a valid json format",
		}
		URLServer(response, req)
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
		body := `{ "url": "" }`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		response := httptest.NewRecorder()
		want := ErrorResponse{
			Error:   "empty URL",
			Code:    "EMPTY_URL",
			Details: "url must not be empty",
		}

		URLServer(response, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)

		var got ErrorResponse

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
