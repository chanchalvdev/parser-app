package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/services"
)

type fakeHealthRepository struct {
	pingErr error
}

func (r *fakeHealthRepository) Ping(_ context.Context) error {
	return r.pingErr
}

func TestHealthEndpointReturnsOk(t *testing.T) {
	t.Run("health", func(t *testing.T) {
		service := services.NewHealthService(&fakeHealthRepository{})
		handler := NewHealthHandler(service)
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		response := httptest.NewRecorder()

		handler.Health().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, response.Code)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body["status"] != "ok" {
			t.Fatalf("expected ok, got %v", body["status"])
		}
	})

	t.Run("ready", func(t *testing.T) {
		service := services.NewHealthService(&fakeHealthRepository{})
		handler := NewHealthHandler(service)
		request := httptest.NewRequest(http.MethodGet, "/ready", nil)
		response := httptest.NewRecorder()

		handler.Ready().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, response.Code)
		}
		var body models.HealthStatus
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Status != "ready" {
			t.Fatalf("expected ready, got %s", body.Status)
		}
	})

	t.Run("not ready", func(t *testing.T) {
		service := services.NewHealthService(&fakeHealthRepository{pingErr: errors.New("db not ready")})
		handler := NewHealthHandler(service)
		request := httptest.NewRequest(http.MethodGet, "/ready", nil)
		response := httptest.NewRecorder()

		handler.Ready().ServeHTTP(response, request)

		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, response.Code)
		}
		payload, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		if len(payload) == 0 {
			t.Fatal("expected json payload")
		}
	})
}

