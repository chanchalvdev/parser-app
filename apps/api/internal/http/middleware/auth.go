package middleware

import (
	"net/http"
	"strings"
	"encoding/json"
)

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	normalized := make([]string, 0, len(roles))
	for _, role := range roles {
		value := strings.ToLower(strings.TrimSpace(role))
		if value != "" {
			normalized = append(normalized, value)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !AuthEnabled(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}

			if HasRole(r.Context(), normalized...) {
				next.ServeHTTP(w, r)
				return
			}

			writeForbidden(w, "forbidden: insufficient permissions")
		})
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	doc := map[string]any{"error": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(doc)
}
