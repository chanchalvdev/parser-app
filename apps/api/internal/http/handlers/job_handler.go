package handlers

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/enterprise-file-platform/api/internal/repositories"
	"github.com/enterprise-file-platform/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type JobHandler struct {
	jobService *services.JobService
}

func NewJobHandler(service *services.JobService) *JobHandler {
	return &JobHandler{jobService: service}
}

func (h *JobHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := h.jobService.ListJobs(r.Context(), services.ListJobsRequest{
			TenantID: r.URL.Query().Get("tenant_id"),
			Page:     parsePageParam(r, "page"),
			PageSize: parsePageSizeParam(r, "page_size"),
		})
		if err != nil {
			writeJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *JobHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "job_id")
		job, err := h.jobService.GetJob(r.Context(), jobID)
		if err != nil {
			writeJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}

func (h *JobHandler) Events() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "job_id")
		response, err := h.jobService.ListJobEvents(r.Context(), services.ListJobEventsRequest{
			JobID:   jobID,
			Page:     parsePageParam(r, "page"),
			PageSize: parsePageSizeParam(r, "page_size"),
		})
		if err != nil {
			writeJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *JobHandler) Retry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "job_id")
		response, err := h.jobService.RetryJob(r.Context(), jobID, services.RetryJobRequest{
			IPAddress: requestIPAddress(r),
			UserAgent: requestUserAgent(r),
		})
		if err != nil {
			writeJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func writeJobError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, nil)
	case errors.Is(err, repositories.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, services.ErrInvalidJobRequest):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func requestIPAddress(r *http.Request) *string {
	rawIP := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if rawIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			rawIP = r.RemoteAddr
		} else {
			rawIP = host
		}
	}
	if rawIP == "" {
		return nil
	}
	if strings.Contains(rawIP, ",") {
		rawIP = strings.TrimSpace(strings.Split(rawIP, ",")[0])
	}
	return &rawIP
}

func requestUserAgent(r *http.Request) *string {
	userAgent := strings.TrimSpace(r.Header.Get("User-Agent"))
	if userAgent == "" {
		return nil
	}
	return &userAgent
}
