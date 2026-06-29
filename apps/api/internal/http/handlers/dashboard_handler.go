package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/enterprise-file-platform/api/internal/services"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: service}
}

func (h *DashboardHandler) Summary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		response, err := h.dashboardService.GetSummary(r.Context(), services.DashboardSummaryRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *DashboardHandler) FileTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		response, err := h.dashboardService.GetFileTypeDistribution(r.Context(), services.DashboardDistributionRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
			Limit:    parseDashboardInt(r.URL.Query().Get("limit"), defaultDashboardLimit()),
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *DashboardHandler) ProcessingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		response, err := h.dashboardService.GetProcessingStatusDistribution(r.Context(), services.DashboardDistributionRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
			Limit:    parseDashboardInt(r.URL.Query().Get("limit"), defaultDashboardLimit()),
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *DashboardHandler) UploadVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		days := parseDashboardInt(r.URL.Query().Get("days"), defaultUploadVolumeDays())

		response, err := h.dashboardService.GetUploadVolume(r.Context(), services.DashboardUploadVolumeRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
			Days:    days,
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *DashboardHandler) ErrorBreakdown() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		response, err := h.dashboardService.GetErrorBreakdown(r.Context(), services.DashboardDistributionRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
			Limit:    parseDashboardInt(r.URL.Query().Get("limit"), defaultDashboardLimit()),
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *DashboardHandler) Entities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		response, err := h.dashboardService.GetTopEntities(r.Context(), services.DashboardDistributionRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
			Limit:    parseDashboardInt(r.URL.Query().Get("limit"), defaultDashboardLimit()),
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *DashboardHandler) ProcessingDuration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.dashboardService == nil {
			writeDashboardError(w, fmt.Errorf("dashboard service is not available"))
			return
		}
		response, err := h.dashboardService.GetProcessingDuration(r.Context(), services.DashboardSummaryRequest{
			TenantID: strings.TrimSpace(r.URL.Query().Get("tenant_id")),
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func writeDashboardError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, nil)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func parseDashboardInt(raw string, defaultValue int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func defaultDashboardLimit() int {
	return 25
}

func defaultUploadVolumeDays() int {
	return 7
}
