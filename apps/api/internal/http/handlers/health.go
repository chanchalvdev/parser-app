package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/services"
)

type HealthHandler struct {
	healthService *services.HealthService
}

func NewHealthHandler(svc *services.HealthService) *HealthHandler {
	return &HealthHandler{healthService: svc}
}

func (h *HealthHandler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func (h *HealthHandler) Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.healthService.Ready(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, models.HealthStatus{
				Status: "not ready",
				Error:  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, models.HealthStatus{Status: "ready"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	writeJSON(w, status, map[string]any{
		"error": msg,
	})
}
