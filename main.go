package main

import (
	"log"
	"net/http"
)

func main() {

	// Create handle func and register route
	http.HandleFunc("/health", HealthCheck)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
