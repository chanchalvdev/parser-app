package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := normalizeOrigins(allowedOrigins)
	if len(allowed) == 0 {
		allowed = map[string]struct{}{ "http://localhost:5173": {} }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setAllowedOrigin(w, r, allowed)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Authorization, Accept, Origin, X-Requested-With, X-Request-ID, X-User-ID, X-User-Roles, X-Tenant-ID",
			)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Add("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func setAllowedOrigin(w http.ResponseWriter, r *http.Request, allowed map[string]struct{}) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		for configured := range allowed {
			w.Header().Set("Access-Control-Allow-Origin", configured)
			return
		}
		return
	}

	if _, ok := allowed[origin]; ok {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
}

func normalizeOrigins(raw []string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		for _, item := range strings.Split(value, ",") {
			origin := strings.TrimSpace(item)
			if origin == "" {
				continue
			}
			allowed[origin] = struct{}{}
		}
	}
	return allowed
}
