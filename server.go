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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"})
		return
	}
	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "empty URL"})
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(URLShortResponse{Short: "abc123"})
}
