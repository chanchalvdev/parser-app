package handlers

import (
	"net/http"

	"github.com/enterprise-file-platform/api/internal/services"
)

type VersionHandler struct {
	service *services.VersionService
	env     string
}

func NewVersionHandler(service *services.VersionService, env string) *VersionHandler {
	return &VersionHandler{service: service, env: env}
}

func (h *VersionHandler) Version() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, h.service.GetVersion(h.env))
	}
}
