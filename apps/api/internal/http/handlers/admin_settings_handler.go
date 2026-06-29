package handlers

import (
	"errors"
	"net/http"

	"github.com/enterprise-file-platform/api/internal/services"
	"github.com/enterprise-file-platform/api/internal/http/middleware"
)

type AdminSettingsHandler struct {
	adminSettingsService *services.AdminSettingsService
}

func NewAdminSettingsHandler(service *services.AdminSettingsService) *AdminSettingsHandler {
	return &AdminSettingsHandler{adminSettingsService: service}
}

func (h *AdminSettingsHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.adminSettingsService == nil {
			writeAdminSettingsError(w, errors.New("admin settings service is not initialized"))
			return
		}

		response, err := h.adminSettingsService.GetSettings(r.Context(), r.URL.Query().Get("tenant_id"))
		if err != nil {
			writeAdminSettingsError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func (h *AdminSettingsHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.adminSettingsService == nil {
			writeAdminSettingsError(w, errors.New("admin settings service is not initialized"))
			return
		}

		var req services.UpdateAdminSettingsRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		response, err := h.adminSettingsService.UpdateSettings(
			r.Context(),
			r.URL.Query().Get("tenant_id"),
			req,
			middleware.ActorUserIDFromContext(r.Context()),
			middleware.RequestIPAddressFromContext(r.Context()),
			middleware.RequestUserAgentFromContext(r.Context()),
		)
		if err != nil {
			writeAdminSettingsError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeAdminSettingsError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, nil)
	case errors.Is(err, services.ErrInvalidSettingsRequest):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
