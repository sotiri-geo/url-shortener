package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/handler"
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

// TODO: change FixedResponse to RepeatResponse. And have a FixedResponse being a config param
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

func TestShortener(t *testing.T) {
	cases := []struct {
		name             string
		payload          string
		setupStore       func(f *FakeStore)
		wantShortCode    string
		wantStatus       int
		wantContentType  string
		wantErrorMessage string
	}{
		{
			name:            "POST /shorten returns a shortened url",
			payload:         `{ "url": "https://example.com" }`,
			setupStore:      func(s *FakeStore) {},
			wantShortCode:   "abc123",
			wantStatus:      http.StatusCreated,
			wantContentType: handler.JsonContentType,
		},
		{
			name:             "request body missing url key",
			payload:          `{ invalid json }`,
			setupStore:       func(f *FakeStore) {},
			wantContentType:  handler.JsonContentType,
			wantStatus:       http.StatusBadRequest,
			wantErrorMessage: handler.ERR_INVALID_JSON,
		},
		{
			name:             "request body with empty string as url",
			payload:          `{"url": ""}`,
			setupStore:       func(f *FakeStore) {},
			wantContentType:  handler.JsonContentType,
			wantStatus:       http.StatusBadRequest,
			wantErrorMessage: handler.ERR_EMPTY_URL,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			store := NewFakeStore()
			tt.setupStore(store)
			gen := NewStubGenerator()
			server := handler.NewShortener(store, gen)

			// execute
			request := newShortenRequest(tt.payload)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			// assert
			assertStatusCode(t, response.Code, tt.wantStatus)
			assertContentType(t, response.Result().Header.Get("content-type"), tt.wantContentType)

			if tt.wantShortCode != "" {
				got, err := getShortCode(response.Body)
				assertNoErr(t, err) // decoding error
				assertShortCode(t, got.Short, tt.wantShortCode)
			}

			if tt.wantErrorMessage != "" {
				got, err := getErrorResponse(response.Body)
				assertNoErr(t, err) // decoding error
				assertErrMessage(t, got.Error, tt.wantErrorMessage)
			}
		})
	}
}

func TestShortenerWithGenerator(t *testing.T) {
	cases := []struct {
		name             string
		payload          string
		setupStore       func(f *FakeStore)
		setupGen         func(g *StubGenerator)
		wantStatus       int
		wantContentType  string
		wantErrorMessage string
		wantGenCallCount int
	}{
		{
			name:    "short code collision requires a retry",
			payload: `{"url": "https://example.com"}`,
			setupStore: func(f *FakeStore) {
				f.urls = map[string]string{"xyz123": "https://test.com"}
			},
			setupGen: func(g *StubGenerator) {
				g.FixedResponse = "xyz123"
				g.GenerateCallCount = 2
			},
			wantStatus:       http.StatusCreated,
			wantContentType:  handler.JsonContentType,
			wantGenCallCount: 2,
		},
	}

	for _, tt := range cases {
		// setup
		store := NewFakeStore()
		tt.setupStore(store)
		gen := NewStubGenerator()
		tt.setupGen(gen)
		server := handler.NewShortener(store, gen)

		// execute
		request := newShortenRequest(tt.payload)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		// assert
		assertStatusCode(t, response.Code, tt.wantStatus)
		assertContentType(t, response.Result().Header.Get("content-type"), tt.wantContentType)

		if tt.wantGenCallCount != 0 {
			assertGenerateCallCount(t, gen.GenerateCallCount, tt.wantGenCallCount)
		}

		if tt.wantErrorMessage != "" {
			got, err := getErrorResponse(response.Body)
			assertNoErr(t, err)
			assertErrMessage(t, got.Error, tt.wantErrorMessage)
		}
	}
}

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("incorrect status code: got %d, want %d", got, want)
	}
}

func assertShortCode(t testing.TB, got, want string) {
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

func assertNoErr(t testing.TB, err error) {
	if err != nil {
		t.Fatalf("should not error: %v", err)
	}
}

func assertGenerateCallCount(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("generate call count: got %d, want %d", got, want)
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

func getErrorResponse(body *bytes.Buffer) (*handler.ErrorResponse, error) {
	var got handler.ErrorResponse

	err := json.NewDecoder(body).Decode(&got)

	if err != nil {
		return nil, err
	}

	return &got, nil
}
