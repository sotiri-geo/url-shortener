package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/sotiri-geo/url-shortener/internal/generator"
	"github.com/sotiri-geo/url-shortener/internal/storage"
)

const (
	ERR_INVALID_JSON                 = "invalid JSON"
	ERR_INVALID_JSON_CODE            = "INVALID_JSON"
	ERR_INVALID_JSON_DETAILS         = "not a valid json format"
	ERR_EMPTY_URL                    = "url must not be empty"
	ERR_EMPTY_URL_CODE               = "EMPTY_URL"
	ERR_EMPTY_URL_DETAILS            = "url must not be empty"
	ERR_SHORT_CODE_NOT_FOUND         = "short code not found"
	ERR_SHORT_CODE_NOT_FOUND_CODE    = "NOT_FOUND"
	ERR_SHORT_CODE_NOT_FOUND_DETAILS = "cannot process redirect without exisiting short code"
	JsonContentType                  = "application/json"
)

var ErrRetryAttemptsExceeded = errors.New("Exhausted retries.")

type URLShortResponse struct {
	Short string
}

type URLRequest struct {
	URL string `json:"url"`
}

type Shortener struct {
	store     storage.URLStore
	generator generator.Generator
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`    // optional
	Details string `json:"details,omitempty"` // optional
	Status  int    `json:"status"`
}

func NewShortener(store storage.URLStore, generator generator.Generator) *Shortener {
	return &Shortener{store, generator}
}

func (e *ErrorResponse) WriteError(w http.ResponseWriter) {
	w.WriteHeader(e.Status)
	json.NewEncoder(w).Encode(e)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

// Implement the Handler interface
func (u *Shortener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", JsonContentType)
	switch r.Method {
	case http.MethodPost:
		u.processURL(w, r)
		return
	}
}

func (u *Shortener) processURL(w http.ResponseWriter, r *http.Request) {
	var req URLRequest
	// try decode the body
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		errResponse := NewErrorResponse(http.StatusBadRequest, ERR_INVALID_JSON, ERR_INVALID_JSON_CODE, ERR_INVALID_JSON_DETAILS)
		errResponse.WriteError(w)
		return
	}
	if req.URL == "" {
		errResponse := NewErrorResponse(http.StatusBadRequest, ERR_EMPTY_URL, ERR_EMPTY_URL_CODE, ERR_EMPTY_URL_DETAILS)
		errResponse.WriteError(w)
		return
	}

	shortCode, err := u.retryShortCode(3) // TODO: pass in rety count as field on struct

	if err != nil {
		errResponse := NewErrorResponse(http.StatusBadRequest, err.Error(), "RETRY_FAIL", fmt.Sprintf("attempted %d retries", 3))
		errResponse.WriteError(w)
		return
	}

	u.store.Save(shortCode, req.URL)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(URLShortResponse{Short: shortCode})
}

func (u *Shortener) retryShortCode(count int) (string, error) {
	shortCode := u.generator.Generate()
	if !u.store.Exists(shortCode) {
		return shortCode, nil
	} else if count > 0 {
		return u.retryShortCode(count - 1)
	}
	return "", ErrRetryAttemptsExceeded
}

func NewErrorResponse(status int, message, code, details string) *ErrorResponse {
	return &ErrorResponse{message, code, details, status}
}
