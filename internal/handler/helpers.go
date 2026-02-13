package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// readJSON decodes a JSON request body into the given value.
func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}

// getQueryParam returns a query parameter value or empty string.
func getQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// getPathParam returns a URL path parameter from chi router.
func getPathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// parseIntParam parses an integer from a string, returning 0 on error.
func parseIntParam(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// parseInt64Param parses an int64 from a string, returning 0 on error.
func parseInt64Param(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
