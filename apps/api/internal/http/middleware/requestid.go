package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if rid == "" {
			rid = uuid.NewString()
		}

		tr := r.WithContext(context.WithValue(r.Context(), requestIDKey, rid))
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, tr)
	})
}

func RequestIDFromContext(ctx context.Context) string {
	value := ctx.Value(requestIDKey)
	if value == nil {
		return ""
	}
	if rid, ok := value.(string); ok {
		return rid
	}
	return ""
}
