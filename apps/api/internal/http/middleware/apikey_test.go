package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAPIKey(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := RequireAPIKey("s3cret")(ok)

	cases := []struct {
		name, method, path, key string
		want                    int
	}{
		{"missing key", http.MethodGet, "/api/v1/search", "", http.StatusUnauthorized},
		{"wrong key", http.MethodGet, "/api/v1/search", "nope", http.StatusUnauthorized},
		{"right key", http.MethodGet, "/api/v1/search", "s3cret", http.StatusOK},
		{"health open", http.MethodGet, "/health", "", http.StatusOK},
		{"docs open", http.MethodGet, "/docs", "", http.StatusOK},
		{"preflight open", http.MethodOptions, "/api/v1/uploads/initiate", "", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.key != "" {
				req.Header.Set("X-API-Key", tc.key)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("got %d want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestRequireAPIKeyEmptyDisables(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	rec := httptest.NewRecorder()
	RequireAPIKey("")(ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/search", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}
