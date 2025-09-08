package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// type HandlerFunc func(ResponseWriter, *Request)

func TestHealthCheckEndpoint(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	HealthCheck(response, req)

	got := response.Body.String()

	want := "ok"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
