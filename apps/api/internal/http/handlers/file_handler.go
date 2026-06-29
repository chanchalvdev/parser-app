package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/enterprise-file-platform/api/internal/repositories"
	"github.com/enterprise-file-platform/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type FileHandler struct {
	fileService *services.FileService
}

func NewFileHandler(service *services.FileService) *FileHandler {
	return &FileHandler{fileService: service}
}

func (h *FileHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := h.fileService.ListFiles(r.Context(), services.ListFilesRequest{
			TenantID:         r.URL.Query().Get("tenant_id"),
			ProcessingStatus: firstNonEmpty(r.URL.Query().Get("status"), r.URL.Query().Get("processing_status")),
			Extension:        r.URL.Query().Get("extension"),
			DetectedFileType: r.URL.Query().Get("detected_file_type"),
			Page:             parsePageParam(r, "page"),
			PageSize:         parsePageSizeParam(r, "page_size"),
		})
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *FileHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "file_id")
		file, err := h.fileService.GetFile(r.Context(), fileID)
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, file)
	}
}

func (h *FileHandler) Children() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "file_id")
		response, err := h.fileService.ListFileChildren(r.Context(), services.ListFileChildrenRequest{
			FileID:   fileID,
			Page:     parsePageParam(r, "page"),
			PageSize: parsePageSizeParam(r, "page_size"),
		})
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *FileHandler) Tree() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "file_id")
		tree, err := h.fileService.GetFileTree(r.Context(), fileID)
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tree)
	}
}

func (h *FileHandler) Records() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "file_id")
		response, err := h.fileService.ListFileRecords(r.Context(), services.ListFileRecordsRequest{
			FileID:   fileID,
			Page:     parsePageParam(r, "page"),
			PageSize: parsePageSizeParam(r, "page_size"),
		})
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *FileHandler) SubmitPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := chi.URLParam(r, "file_id")

		var req services.SubmitFilePasswordRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		response, err := h.fileService.SubmitFilePassword(
			r.Context(),
			fileID,
			req,
			services.SubmitFilePasswordContext{
				IPAddress: requestIPAddress(r),
				UserAgent: requestUserAgent(r),
			},
		)
		if err != nil {
			writeFileError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeFileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, services.ErrInvalidFileRequest):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, services.ErrInvalidFilePasswordRequest):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func parsePageParam(r *http.Request, key string) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0
	}
	return value
}

func parsePageSizeParam(r *http.Request, key string) int {
	size := parsePageParam(r, key)
	if size != 0 {
		return size
	}
	legacy := parsePageParam(r, "limit")
	if legacy > 0 {
		return legacy
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
