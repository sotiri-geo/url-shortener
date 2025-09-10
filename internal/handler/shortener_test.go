package handler_test

import (
	"bytes"
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

func (f *FakeStore) Save(shortCode, original string) error {
	f.urls[shortCode] = original
	return nil
}

func (f *FakeStore) Exists(shortCode string) bool {
	_, exists := f.urls[shortCode]
	return exists
}

func NewFakeStore() *FakeStore {
	return &FakeStore{urls: make(map[string]string)}
}

func NewFakeStoreWithUrls(urls map[string]string) *FakeStore {
	return &FakeStore{urls: urls}
}

type StubGenerator struct {
	GenerateCallCount int
	FixedResponse     string
	Repeat            int
}

func (s *StubGenerator) Generate() string {
	// Forcing a collision
	if s.Repeat > 0 {
		s.GenerateCallCount++
		s.Repeat--
		return s.FixedResponse
	}
	return "abc123"
}

func NewStubGenerator() *StubGenerator {
	return &StubGenerator{}
}

func NewStubGeneratorWithFixedResponse(fixedResponse string, repeat int) *StubGenerator {
	return &StubGenerator{FixedResponse: fixedResponse, Repeat: repeat}
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
		req := newShortenRequest(`{ "url": "https://example.com" }`)
		store := NewFakeStore()
		server := handler.NewShortener(store, NewStubGenerator())
		response := httptest.NewRecorder()
		want := "abc123"
		server.ServeHTTP(response, req)

		// decode
		got, err := getShortCode(response.Body)

		if err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		assertStatusCode(t, response.Code, http.StatusCreated)
		assertContentType(t, response.Result().Header.Get("content-type"), "application/json")
		assertURL(t, got.Short, want)
	})

	t.Run("POST /shorten stores state", func(t *testing.T) {
		req := newShortenRequest(`{ "url": "https://example.com" }`)
		store := NewFakeStore()
		server := handler.NewShortener(store, NewStubGenerator())
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		got, err := getShortCode(response.Body)
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
		server := handler.NewShortener(store, NewStubGenerator())
		req := newShortenRequest(`{ invalid json }`)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)

		// Check error response
		var got handler.ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		assertErrMessage(t, got.Error, handler.ERR_INVALID_JSON)
	})

	t.Run("bad client request with empty url", func(t *testing.T) {
		store := NewFakeStore()
		server := handler.NewShortener(store, NewStubGenerator())
		req := newShortenRequest(`{ "url": "" }`)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)

		var got handler.ErrorResponse

		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

		assertErrMessage(t, got.Error, handler.ERR_EMPTY_URL)
	})

	t.Run("generated code existing requires a retry", func(t *testing.T) {
		store := NewFakeStoreWithUrls(map[string]string{"xyz123": "https://test.com"})
		wantCallCount := 2
		gen := NewStubGeneratorWithFixedResponse("xyz123", wantCallCount)
		server := handler.NewShortener(store, gen)
		req := newShortenRequest(`{"url": "https://example.com"}`)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)

		// Generator should be called multiple times on collision

		if gen.GenerateCallCount != wantCallCount {
			t.Errorf("generator should have retried: call count found %d", gen.GenerateCallCount)
		}
	})

	t.Run("exceeded retry attempts", func(t *testing.T) {
		store := NewFakeStoreWithUrls(map[string]string{"xyz123": "https://test.com"})
		gen := NewStubGeneratorWithFixedResponse("xyz123", 5) // repeats more times
		server := handler.NewShortener(store, gen)
		req := newShortenRequest(`{"url": "https://example.com"}`)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)

		// Should exceed count returning an error
		got, err := getErrorResponse(response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		assertStatusCode(t, response.Code, http.StatusInternalServerError)
		wantErrMessage := handler.ErrRetryAttemptsExceeded.Error()
		assertErrMessage(t, got.Error, wantErrMessage)

	})
}

func getErrorResponse(response *httptest.ResponseRecorder) (*handler.ErrorResponse, error) {
	var got handler.ErrorResponse

	err := json.NewDecoder(response.Body).Decode(&got)

	if err != nil {
		return &got, err
	}

	return &got, nil
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

func assertErrMessage(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got error message %q, want %q", got, want)
	}
}

func assertContentType(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("incorrect content type: got %q, want %q", got, want)
	}
}

func newShortenRequest(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
}

func getShortCode(body *bytes.Buffer) (*handler.URLShortResponse, error) {
	var got handler.URLShortResponse

	// decode
	err := json.NewDecoder(body).Decode(&got)

	if err != nil {
		return nil, err
	}

	return &got, nil
}
