package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type URLShortResponse struct {
	Short string
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

func URLServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(URLShortResponse{Short: "abc123"})
}
