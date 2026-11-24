package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// RequireAPIKey rejects /api/v1 requests whose X-API-Key header does not match key.
// /health, /ready and /docs stay open so platform probes keep working. CORS preflights
// (OPTIONS) are answered by the CORS middleware before reaching this one.
//
// An empty key disables the check (local development).
func RequireAPIKey(key string) func(http.Handler) http.Handler {
	expected := []byte(strings.TrimSpace(key))
	return func(next http.Handler) http.Handler {
		if len(expected) == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions || !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}
			provided := []byte(strings.TrimSpace(r.Header.Get("X-API-Key")))
			if subtle.ConstantTimeCompare(provided, expected) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"missing or invalid X-API-Key"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
