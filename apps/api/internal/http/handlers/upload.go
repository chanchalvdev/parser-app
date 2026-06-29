package handlers

import (
	"encoding/json"
	"errors"
	"context"
	"io"
	"net/http"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/repositories"
	"github.com/enterprise-file-platform/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type uploadService interface {
	InitiateUpload(ctx context.Context, request services.UploadInitiateRequest) (*services.UploadInitiateResponse, error)
	CompleteUpload(ctx context.Context, request services.UploadCompleteRequest) (*services.UploadCompleteResponse, error)
	GetUpload(ctx context.Context, uploadID string) (*models.Upload, error)
}

type UploadHandler struct {
	uploadService uploadService
}

func NewUploadHandler(service uploadService) *UploadHandler {
	return &UploadHandler{uploadService: service}
}

func (h *UploadHandler) InitiateUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.UploadInitiateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		response, err := h.uploadService.InitiateUpload(r.Context(), req)
		if err != nil {
			writeUploadError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, response)
	}
}

func (h *UploadHandler) CompleteUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.UploadCompleteRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		response, err := h.uploadService.CompleteUpload(r.Context(), req)
		if err != nil {
			writeUploadError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, response)
	}
}

func (h *UploadHandler) GetUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uploadID := chi.URLParam(r, "upload_id")
		upload, err := h.uploadService.GetUpload(r.Context(), uploadID)
		if err != nil {
			writeUploadError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, upload)
	}
}

func writeUploadError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, services.ErrInvalidUploadRequest):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, services.ErrUploadTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func decodeJSON(r *http.Request, payload any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(payload); err != nil {
		return err
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return ErrInvalidUploadBody
	}
	_ = r.Body.Close()
	return nil
}

var ErrInvalidUploadBody = errors.New("invalid JSON payload")
