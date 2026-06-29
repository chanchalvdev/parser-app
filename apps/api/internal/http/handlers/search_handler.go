package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/enterprise-file-platform/api/internal/services"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(service *services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: service}
}

func (h *SearchHandler) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.SearchRecordsRequest
		switch r.Method {
		case http.MethodGet:
			req = parseSearchQuery(r)
		case http.MethodPost:
			if err := decodeJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
		default:
			writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
			return
		}

		response, err := h.searchService.Search(r.Context(), req)
		if err != nil {
			writeSearchError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *SearchHandler) Suggestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
			return
		}
		req := parseSearchQuery(r)
		response, err := h.searchService.Suggestions(r.Context(), req)
		if err != nil {
			writeSearchError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func parseSearchQuery(r *http.Request) services.SearchRecordsRequest {
	q := r.URL.Query()
	page := 1
	if raw := strings.TrimSpace(q.Get("page")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			page = value
		}
	}

	pageSize := 0
	if raw := strings.TrimSpace(q.Get("page_size")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			pageSize = value
		}
	}
	if pageSize == 0 {
		if raw := strings.TrimSpace(q.Get("limit")); raw != "" {
			if value, err := strconv.Atoi(raw); err == nil {
				pageSize = value
			}
		}
	}

	return services.SearchRecordsRequest{
		Q:                q.Get("q"),
		FileID:           q.Get("file_id"),
		Extension:        q.Get("extension"),
		DetectedFileType: q.Get("detected_file_type"),
		RecordType:       q.Get("record_type"),
		DateFrom:         q.Get("date_from"),
		DateTo:           q.Get("date_to"),
		IP:               q.Get("ip"),
		Email:            q.Get("email"),
		Domain:           q.Get("domain"),
		JobID:            q.Get("job_id"),
		Page:             page,
		PageSize:         pageSize,
		Sort:             q.Get("sort"),
		TenantID:         q.Get("tenant_id"),
	}
}

func writeSearchError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, nil)
	case errors.Is(err, services.ErrInvalidSearchRequest):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
