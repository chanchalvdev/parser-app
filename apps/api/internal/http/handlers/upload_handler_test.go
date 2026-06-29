package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/services"
)

type fakeUploadService struct {
	initiateFn func(context.Context, services.UploadInitiateRequest) (*services.UploadInitiateResponse, error)
	completeFn  func(context.Context, services.UploadCompleteRequest) (*services.UploadCompleteResponse, error)
	getUploadFn func(context.Context, string) (*models.Upload, error)
}

func (f fakeUploadService) InitiateUpload(ctx context.Context, req services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
	return f.initiateFn(ctx, req)
}

func (f fakeUploadService) CompleteUpload(ctx context.Context, req services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
	return f.completeFn(ctx, req)
}

func (f fakeUploadService) GetUpload(ctx context.Context, uploadID string) (*models.Upload, error) {
	return f.getUploadFn(ctx, uploadID)
}

func TestUploadInitiateHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var got services.UploadInitiateRequest
		handler := NewUploadHandler(fakeUploadService{
			initiateFn: func(_ context.Context, req services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
				got = req
				return &services.UploadInitiateResponse{
					UploadID: "upload-1",
					ObjectKey: "raw/file.txt",
					UploadURL: "https://storage.local/upload",
					ExpiresInSeconds: 60,
				}, nil
			},
			completeFn: func(_ context.Context, _ services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
				t.Fatal("complete should not be called")
				return nil, nil
			},
			getUploadFn: func(context.Context, string) (*models.Upload, error) {
				t.Fatal("getUpload should not be called")
				return nil, nil
			},
		})

		request := httptest.NewRequest(http.MethodPost, "/uploads/initiate", strings.NewReader(`{"file_name":"report.csv","content_type":"text/csv","size_bytes":1024}`))
		response := httptest.NewRecorder()

		handler.InitiateUpload().ServeHTTP(response, request)

		if response.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, response.Code)
		}
		if got.FileName != "report.csv" {
			t.Fatalf("expected file_name report.csv, got %s", got.FileName)
		}
		if got.SizeBytes != 1024 {
			t.Fatalf("expected size 1024, got %d", got.SizeBytes)
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		handler := NewUploadHandler(fakeUploadService{
			initiateFn: func(context.Context, services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
				t.Fatal("initiate should not be called")
				return nil, nil
			},
			completeFn: func(context.Context, services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
				return nil, nil
			},
			getUploadFn: func(context.Context, string) (*models.Upload, error) {
				return nil, nil
			},
		})
		request := httptest.NewRequest(http.MethodPost, "/uploads/initiate", strings.NewReader(`{"file_name":`))
		response := httptest.NewRecorder()

		handler.InitiateUpload().ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("service too large error", func(t *testing.T) {
		handler := NewUploadHandler(fakeUploadService{
			initiateFn: func(context.Context, services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
				return nil, services.ErrUploadTooLarge
			},
			completeFn: func(context.Context, services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
				return nil, nil
			},
			getUploadFn: func(context.Context, string) (*models.Upload, error) {
				return nil, nil
			},
		})
		request := httptest.NewRequest(http.MethodPost, "/uploads/initiate", strings.NewReader(`{"file_name":"report.csv","size_bytes":100}`))
		response := httptest.NewRecorder()

		handler.InitiateUpload().ServeHTTP(response, request)

		if response.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected %d, got %d", http.StatusRequestEntityTooLarge, response.Code)
		}
	})
}

func TestUploadCompleteHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := NewUploadHandler(fakeUploadService{
			initiateFn: func(context.Context, services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
				return nil, nil
			},
			completeFn: func(_ context.Context, req services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
				if req.UploadID != "upload-1" {
					t.Fatalf("unexpected upload_id: %s", req.UploadID)
				}
				return &services.UploadCompleteResponse{
					UploadID: "upload-1",
					FileID:   "file-1",
					JobID:    "job-1",
					Status:   "queued",
				}, nil
			},
			getUploadFn: func(context.Context, string) (*models.Upload, error) {
				return nil, nil
			},
		})

		request := httptest.NewRequest(http.MethodPost, "/uploads/complete", strings.NewReader(`{"upload_id":"upload-1"}`))
		response := httptest.NewRecorder()

		handler.CompleteUpload().ServeHTTP(response, request)

		if response.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, response.Code)
		}
		var payload services.UploadCompleteResponse
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.UploadID != "upload-1" {
			t.Fatalf("expected upload-1, got %s", payload.UploadID)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		var called bool
		handler := NewUploadHandler(fakeUploadService{
			initiateFn: func(context.Context, services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
				return nil, nil
			},
			completeFn: func(context.Context, services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
				called = true
				return nil, nil
			},
			getUploadFn: func(context.Context, string) (*models.Upload, error) {
				return nil, nil
			},
		})
		request := httptest.NewRequest(http.MethodPost, "/uploads/complete", strings.NewReader(`not-json`))
		response := httptest.NewRecorder()

		handler.CompleteUpload().ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, response.Code)
		}
		if called {
			t.Fatal("complete handler should not be called for invalid request")
		}
	})

	t.Run("bad service result", func(t *testing.T) {
		handler := NewUploadHandler(fakeUploadService{
			initiateFn: func(context.Context, services.UploadInitiateRequest) (*services.UploadInitiateResponse, error) {
				return nil, nil
			},
			completeFn: func(context.Context, services.UploadCompleteRequest) (*services.UploadCompleteResponse, error) {
				return nil, fmt.Errorf("%w: upload_id is required", services.ErrInvalidUploadRequest)
			},
			getUploadFn: func(context.Context, string) (*models.Upload, error) {
				return nil, nil
			},
		})
		request := httptest.NewRequest(http.MethodPost, "/uploads/complete", strings.NewReader(`{"upload_id":""}`))
		response := httptest.NewRecorder()

		handler.CompleteUpload().ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, response.Code)
		}
	})
}
