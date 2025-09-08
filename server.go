package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type URLShortResponse struct {
	Short string
}

type URLRequest struct {
	URL string `json:"url"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`    // optional
	Details string `json:"details,omitempty"` // optional
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

func URLServer(w http.ResponseWriter, r *http.Request) {
	var req URLRequest
	// try decode the body
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeErrorResponse(w, "invalid JSON", "INVALID_JSON", "not a valid json format")
		return
	}
	if req.URL == "" {
		writeErrorResponse(w, "empty URL", "EMPTY_URL", "url must not be empty")
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(URLShortResponse{Short: "abc123"})
}

func writeErrorResponse(w http.ResponseWriter, message, code, details string) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{message, code, details})
}
