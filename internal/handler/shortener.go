package handler

import (
	"encoding/json"
	"io"
	"net/http"

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
	jsonContentType                  = "application/json"
)

type URLShortResponse struct {
	Short string
}

type URLRequest struct {
	URL string `json:"url"`
}

type Shortener struct {
	store storage.URLStore
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`    // optional
	Details string `json:"details,omitempty"` // optional
	Status  int    `json:"status"`
}

func NewShortener(store storage.URLStore) *Shortener {
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
	w.Header().Set("content-type", jsonContentType)
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
	// Write to store
	shortCode := u.store.GetShortURL(req.URL)
	u.store.Save(shortCode, req.URL)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(URLShortResponse{Short: shortCode})
}

func NewErrorResponse(status int, message, code, details string) *ErrorResponse {
	return &ErrorResponse{message, code, details, status}
}
