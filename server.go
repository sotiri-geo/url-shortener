package main

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	ERR_INVALID_JSON         = "invalid JSON"
	ERR_INVALID_JSON_CODE    = "INVALID_JSON"
	ERR_INVALID_JSON_DETAILS = "not a valid json format"
	ERR_EMPTY_URL            = "url must not be empty"
	ERR_EMPTY_URL_CODE       = "EMPTY_URL"
	ERR_EMPTY_URL_DETAILS    = "url must not be empty"
)

type URLShortResponse struct {
	Short string
}

type URLRequest struct {
	URL string `json:"url"`
}

type URLStore interface {
	GetShortURL(url string) string
}

type URLServer struct {
	store URLStore
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`    // optional
	Details string `json:"details,omitempty"` // optional
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

// Implement the Handler interface
func (u *URLServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req URLRequest
	// try decode the body
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeErrorResponse(w, ERR_INVALID_JSON, ERR_INVALID_JSON_CODE, ERR_INVALID_JSON_DETAILS)
		return
	}
	if req.URL == "" {
		writeErrorResponse(w, ERR_EMPTY_URL, ERR_EMPTY_URL_CODE, ERR_EMPTY_URL_DETAILS)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(URLShortResponse{Short: u.store.GetShortURL(req.URL)})
}

func writeErrorResponse(w http.ResponseWriter, message, code, details string) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{message, code, details})
}
