package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"data": v})
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": message})
}

// decodeJSON decodes the request body into v.
func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// readRawBody reads the entire request body as bytes.
func readRawBody(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// parsePositiveInt parses a string as a positive integer, returning fallback on failure.
func parsePositiveInt(s string, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

// isTestFromQuery checks for isTest=true in GET query parameters.
func isTestFromQuery(r *http.Request) bool {
	return r.URL.Query().Get("isTest") == "true"
}
