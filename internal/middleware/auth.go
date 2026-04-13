package middleware

import (
	"net/http"
)

// Auth validates the Bearer token if an API key is configured.
// If apiKey is empty, all requests are allowed through.
func Auth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey != "" && r.Header.Get("Authorization") != "Bearer "+apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
