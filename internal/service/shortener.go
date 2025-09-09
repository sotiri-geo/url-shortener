package service

import (
	"encoding/json"
	"io"
	"net/http"
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
)

type URLShortResponse struct {
	Short string
}

type URLRequest struct {
	URL string `json:"url"`
}

type URLStore interface {
	GetShortURL(url string) string
	GetOriginalURL(shortCode string) (string, bool)
}

type Shortener struct {
	store URLStore
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`    // optional
	Details string `json:"details,omitempty"` // optional
	Status  int    `json:"status"`
}

func NewShortener(store URLStore) *Shortener {
	return &Shortener{store}
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

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(URLShortResponse{Short: u.store.GetShortURL(req.URL)})
}

func NewErrorResponse(status int, message, code, details string) *ErrorResponse {
	return &ErrorResponse{message, code, details, status}
}
