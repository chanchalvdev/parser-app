package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(body)
	rw.size += n
	return n, err
}

func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			if rw.status == 0 {
				rw.status = http.StatusOK
			}

			durationMS := int64(time.Since(start) / time.Millisecond)

			if rw.status >= http.StatusInternalServerError {
				log.Error("request finished with server error",
					zap.String("request_id", RequestIDFromContext(r.Context())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", rw.status),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("query", r.URL.RawQuery),
					zap.Int64("duration_ms", durationMS),
					zap.Int("bytes", rw.size),
				)
				return
			}

			if rw.status >= http.StatusBadRequest {
				log.Warn("request finished with client error",
					zap.String("request_id", RequestIDFromContext(r.Context())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", rw.status),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("query", r.URL.RawQuery),
					zap.Int64("duration_ms", durationMS),
					zap.Int("bytes", rw.size),
				)
				return
			}

			log.Info("request finished",
				zap.String("request_id", RequestIDFromContext(r.Context())),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("query", r.URL.RawQuery),
				zap.Int64("duration_ms", durationMS),
				zap.Int("bytes", rw.size),
			)
		})
	}
}
