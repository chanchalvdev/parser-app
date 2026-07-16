package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/repositories"
	"github.com/enterprise-file-platform/api/internal/http/middleware"
	"github.com/enterprise-file-platform/api/internal/queue"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	defaultTenantID      = "00000000-0000-0000-0000-000000000001"
	defaultUserID        = "11111111-1111-1111-1111-111111111100"
	uploadURLExpiration  = 3600
	uploadURLExpirationD = 3600 * time.Second
	ingestionQueueErrorCode = "QUEUE_ENQUEUE_FAILED"
)

var (
	ErrInvalidUploadRequest  = errors.New("invalid upload request")
	ErrUploadTooLarge       = errors.New("upload exceeds max_upload_size_mb")
	ErrUploadQueueOperation = errors.New("upload queue operation failed")
)

type UploadService struct {
	uploadRepo       repositories.UploadRepository
	fileRepo         repositories.FileRepository
	jobRepo          repositories.JobRepository
	jobEventRepo     repositories.JobEventRepository
	settingsRepo     repositories.SettingsRepository
	auditLogRepo     repositories.AuditLogRepository
	minioClient      *minio.Client
	queueProducer    queue.Producer
	minioBucket      string
	minioPublicURL   string
}

type UploadInitiateRequest struct {
	FileName          string `json:"file_name"`
	ContentType       string `json:"content_type"`
	SizeBytes         int64  `json:"size_bytes"`
	PasswordProvided  bool   `json:"password_provided"`
}

type UploadInitiateResponse struct {
	UploadID          string `json:"upload_id"`
	ObjectKey         string `json:"object_key"`
	UploadURL         string `json:"upload_url"`
	ExpiresInSeconds  int64  `json:"expires_in_seconds"`
}

type UploadCompleteRequest struct {
	UploadID string `json:"upload_id"`
}

type UploadCompleteResponse struct {
	UploadID string `json:"upload_id"`
	FileID   string `json:"file_id"`
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
}

type uploadMetadata struct {
	FileName         string `json:"file_name"`
	ContentType      string `json:"content_type"`
	SizeBytes        int64  `json:"size_bytes"`
	PasswordProvided bool   `json:"password_provided"`
}

func NewUploadService(
	uploadRepo repositories.UploadRepository,
	fileRepo repositories.FileRepository,
	jobRepo repositories.JobRepository,
	jobEventRepo repositories.JobEventRepository,
	settingsRepo repositories.SettingsRepository,
	auditLogRepo repositories.AuditLogRepository,
	minioClient *minio.Client,
	queueProducer queue.Producer,
	minioBucket string,
	minioPublicURL string,
) *UploadService {
	return &UploadService{
		uploadRepo:   uploadRepo,
		fileRepo:     fileRepo,
		jobRepo:      jobRepo,
		jobEventRepo: jobEventRepo,
		settingsRepo: settingsRepo,
		auditLogRepo: auditLogRepo,
		minioClient:  minioClient,
		queueProducer: queueProducer,
		minioBucket:  minioBucket,
		minioPublicURL: minioPublicURL,
	}
}

func (s *UploadService) InitiateUpload(ctx context.Context, req UploadInitiateRequest) (*UploadInitiateResponse, error) {
	if s.minioClient == nil {
		return nil, fmt.Errorf("upload service is not initialized: minio client")
	}
	if s.settingsRepo == nil || s.uploadRepo == nil || s.auditLogRepo == nil {
		return nil, fmt.Errorf("upload service is not initialized")
	}

	fileName := strings.TrimSpace(req.FileName)
	contentType := strings.TrimSpace(req.ContentType)

	if fileName == "" {
		return nil, fmt.Errorf("%w: file_name is required", ErrInvalidUploadRequest)
	}
	if req.SizeBytes <= 0 {
		return nil, fmt.Errorf("%w: size_bytes must be greater than zero", ErrInvalidUploadRequest)
	}

	maxBytes, err := s.maxUploadBytes(ctx)
	if err != nil {
		return nil, err
	}
	if req.SizeBytes > maxBytes {
		return nil, fmt.Errorf("%w: size_bytes=%d > %d", ErrUploadTooLarge, req.SizeBytes, maxBytes)
	}

	normalizedFileName, err := sanitizeFileName(fileName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUploadRequest, err)
	}

	metadata := uploadMetadata{
		FileName:         normalizedFileName,
		ContentType:      contentType,
		SizeBytes:        req.SizeBytes,
		PasswordProvided: req.PasswordProvided,
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("serialize upload metadata: %w", err)
	}

	uploadID := uuid.NewString()
	upload := &models.Upload{
		ID:             uploadID,
		TenantID:       defaultTenantID,
		InitiatedBy:    ptrString(defaultUserID),
		Status:         "pending",
		OriginalSource: ptrString(normalizedFileName),
		StoragePrefix:  ptrString(storagePrefix(defaultTenantID, uploadID)),
		TotalFiles:     1,
		TotalSizeBytes: req.SizeBytes,
		Metadata:       metadataBytes,
	}

	createdUpload, err := s.uploadRepo.CreateUpload(ctx, upload)
	if err != nil {
		return nil, err
	}

	if createdUpload.StoragePrefix == nil || *createdUpload.StoragePrefix == "" {
		return nil, fmt.Errorf("upload storage prefix is empty after create")
	}

	objectKey := storageObjectKey(*createdUpload.StoragePrefix, normalizedFileName)
	presignedURL, err := s.minioClient.PresignedPutObject(
		ctx,
		s.minioBucket,
		objectKey,
		uploadURLExpirationD,
	)
	if err != nil {
		return nil, fmt.Errorf("create presigned upload url: %w", err)
	}
	if s.minioPublicURL != "" {
		publicURL, rewriteErr := rewritePresignedURL(presignedURL.String(), s.minioPublicURL)
		if rewriteErr != nil {
			return nil, fmt.Errorf("rewrite presigned upload url: %w", rewriteErr)
		}
		parsedURL, parseErr := url.Parse(publicURL)
		if parseErr != nil {
			return nil, fmt.Errorf("parse rewritten upload url: %w", parseErr)
		}
		presignedURL = parsedURL
	}

	if err := s.createUploadAuditLog(
		ctx,
		defaultTenantID,
		"upload.initiated",
		map[string]any{
			"upload_id":      createdUpload.ID,
			"file_name":      normalizedFileName,
			"size_bytes":     req.SizeBytes,
			"content_type":   contentType,
			"password_provided": req.PasswordProvided,
			"object_key":     objectKey,
		},
	); err != nil {
		return nil, fmt.Errorf("create upload initiated audit log: %w", err)
	}

	return &UploadInitiateResponse{
		UploadID:          createdUpload.ID,
		ObjectKey:         objectKey,
		UploadURL:         presignedURL.String(),
		ExpiresInSeconds: int64(uploadURLExpiration),
	}, nil
}

func (s *UploadService) CompleteUpload(ctx context.Context, req UploadCompleteRequest) (*UploadCompleteResponse, error) {
	if s.minioClient == nil {
		return nil, fmt.Errorf("upload service is not initialized: minio client")
	}
	if s.queueProducer == nil || s.uploadRepo == nil || s.fileRepo == nil || s.jobRepo == nil || s.jobEventRepo == nil || s.auditLogRepo == nil {
		return nil, fmt.Errorf("upload service is not initialized")
	}

	uploadID := strings.TrimSpace(req.UploadID)
	if uploadID == "" {
		return nil, fmt.Errorf("%w: upload_id is required", ErrInvalidUploadRequest)
	}

	upload, err := s.uploadRepo.GetUploadByID(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	if upload.StoragePrefix == nil || *upload.StoragePrefix == "" {
		return nil, fmt.Errorf("%w: upload storage_prefix missing", ErrInvalidUploadRequest)
	}
	if upload.OriginalSource == nil || *upload.OriginalSource == "" {
		return nil, fmt.Errorf("%w: original upload file_name missing", ErrInvalidUploadRequest)
	}

	meta, err := parseUploadMetadata(upload.Metadata)
	if err != nil {
		if upload.OriginalSource == nil || strings.TrimSpace(*upload.OriginalSource) == "" {
			return nil, fmt.Errorf("parse upload metadata: %w", err)
		}
		meta = uploadMetadata{
			FileName: strings.TrimSpace(*upload.OriginalSource),
		}
	}
	meta.FileName = strings.TrimSpace(meta.FileName)
	meta.FileName, err = sanitizeFileName(meta.FileName)
	if err != nil {
		return nil, fmt.Errorf("%w: file name invalid: %v", ErrInvalidUploadRequest, err)
	}
	if err := validateUploadMetadata(meta); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUploadRequest, err)
	}

	objectKey := storageObjectKey(*upload.StoragePrefix, meta.FileName)
	existingObj, err := s.minioClient.StatObject(ctx, s.minioBucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: object not found in object storage: %s", ErrInvalidUploadRequest, err)
	}

	detectedMimeType := meta.ContentType

	storageSize := existingObj.Size
	if storageSize <= 0 {
		storageSize = upload.TotalSizeBytes
	}

	file := &models.File{
		TenantID:           upload.TenantID,
		ParentFileID:       nil,
		UploadID:           &upload.ID,
		OriginalName:       meta.FileName,
		NormalizedName:     ptrString(strings.ToLower(meta.FileName)),
		Extension:          ptrString(normalizeExtension(meta.FileName)),
		DetectedMimeType:   ptrFromStringValue(detectedMimeType),
		DetectedFileType:   ptrFromStringValue(inferDetectedFileType(meta.FileName, detectedMimeType)),
		StoragePath:        objectKey,
		SizeBytes:          storageSize,
		Sha256Hash:         nil,
		Depth:              0,
		IsArchive:          isArchive(meta.FileName, detectedMimeType),
		IsPasswordProtected: meta.PasswordProvided,
		ProcessingStatus:    "queued",
		CreatedBy:           ptrString(defaultUserID),
	}

	createdFile, err := s.fileRepo.CreateFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("create file record: %w", err)
	}

	job := &models.IngestionJob{
		ID:              uuid.NewString(),
		TenantID:        upload.TenantID,
		RootFileID:      createdFile.ID,
		Status:          "pending",
		CurrentStage:    ptrString("created"),
		ProgressPercent: 0,
		RetryCount:      0,
		ErrorCode:       nil,
		ErrorMessage:    nil,
		StartedAt:       nil,
		CompletedAt:     nil,
	}

	createdJob, err := s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("create ingestion job: %w", err)
	}

	queueMessage := queue.IngestionJobMessage{
		JobID:       createdJob.ID,
		TenantID:    createdJob.TenantID,
		RootFileID:  createdFile.ID,
		StoragePath: createdFile.StoragePath,
		Bucket:      s.minioBucket,
		OriginalName: createdFile.OriginalName,
		Depth:       createdFile.Depth,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.enqueueAndTrack(ctx, createdJob, queueMessage, upload.ID); err != nil {
		return nil, err
	}

	if err := s.jobRepo.UpdateJobStatus(
		ctx,
		createdJob.ID,
		"queued",
		ptrString("queued"),
		ptrFloat64(0),
		nil,
		nil,
		nil,
		nil,
	); err != nil {
		return nil, fmt.Errorf("mark ingestion job as queued: %w", err)
	}

	successEventDetails, err := json.Marshal(map[string]any{
		"upload_id": upload.ID,
		"file_id":   createdFile.ID,
		"job_id":    createdJob.ID,
		"message":   "upload completed and queued for parsing",
		"queued_at": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("build job event details: %w", err)
	}
	if err := s.createJobEvent(
		ctx,
		createdJob.ID,
		createdJob.TenantID,
		"upload.completed",
		"upload completed and queued for parsing",
		json.RawMessage(successEventDetails),
	); err != nil {
		return nil, err
	}

	if err := s.uploadRepo.MarkUploadCompleted(ctx, upload.ID); err != nil {
		return nil, fmt.Errorf("mark upload completed: %w", err)
	}

	if err := s.createUploadAuditLog(
		ctx,
		createdJob.TenantID,
		"upload.completed",
		map[string]any{
			"upload_id":      upload.ID,
			"file_id":        createdFile.ID,
			"job_id":         createdJob.ID,
			"status":         "queued",
			"message":        "upload completed and queued for parsing",
		},
	); err != nil {
		return nil, fmt.Errorf("create upload completed audit log: %w", err)
	}

	return &UploadCompleteResponse{
		UploadID: upload.ID,
		FileID:   createdFile.ID,
		JobID:    createdJob.ID,
		Status:   "queued",
	}, nil
}

func (s *UploadService) enqueueAndTrack(
	ctx context.Context,
	createdJob *models.IngestionJob,
	message queue.IngestionJobMessage,
	uploadID string,
) error {
	if err := s.queueProducer.EnqueueIngestionJob(ctx, message); err != nil {
		now := time.Now().UTC()
		stage := "enqueue_failed"
		errCode := ingestionQueueErrorCode
		errMessage := fmt.Sprintf("failed to enqueue ingestion job: %v", err)
		updateErr := s.jobRepo.UpdateJobStatus(
			ctx,
			createdJob.ID,
			"failed",
			ptrString(stage),
			ptrFloat64(0),
			&errCode,
			&errMessage,
			nil,
			&now,
		)
		if updateErr != nil {
			return fmt.Errorf("mark failed job status: %w", updateErr)
		}

		eventDetails, marshalErr := json.Marshal(map[string]any{
			"upload_id":  uploadID,
			"job_id":     createdJob.ID,
			"file_id":    createdJob.RootFileID,
			"stage":      stage,
			"error":      err.Error(),
			"created_at": time.Now().UTC().Format(time.RFC3339),
		})
		if marshalErr != nil {
			return fmt.Errorf("build queue failure event details: %w", marshalErr)
		}
		if createErr := s.createJobEvent(
			ctx,
			createdJob.ID,
			createdJob.TenantID,
			"upload.queue.failed",
			"failed to enqueue ingestion job",
			json.RawMessage(eventDetails),
		); createErr != nil {
			return fmt.Errorf("create queue failure event: %w", createErr)
		}
		return fmt.Errorf("%w: %v", ErrUploadQueueOperation, err)
	}

	return nil
}

func (s *UploadService) createJobEvent(
	ctx context.Context,
	jobID string,
	tenantID string,
	eventType string,
	eventMessage string,
	eventDetails json.RawMessage,
) error {
	if len(eventDetails) == 0 {
		eventDetails = json.RawMessage("{}")
	}

	evtMessage := eventMessage
	_, err := s.jobEventRepo.CreateJobEvent(ctx, &models.JobEvent{
		TenantID:     tenantID,
		JobID:        jobID,
		EventType:    eventType,
		EventMessage: &evtMessage,
		EventDetails: eventDetails,
		CreatedBy:    ptrString(defaultUserID),
	})
	if err != nil {
		return fmt.Errorf("create job event: %w", err)
	}
	return nil
}

func (s *UploadService) createUploadAuditLog(
	ctx context.Context,
	tenantID string,
	action string,
	details map[string]any,
) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	_, err = s.auditLogRepo.CreateAuditLog(ctx, &models.AuditLog{
		TenantID:   tenantID,
		ActorUserID: middleware.ActorUserIDFromContext(ctx),
		Action:     action,
		EntityType: ptrString("upload"),
		Details:    detailsJSON,
		IPAddress:  middleware.RequestIPAddressFromContext(ctx),
		UserAgent:  middleware.RequestUserAgentFromContext(ctx),
	})
	return err
}

func (s *UploadService) GetUpload(ctx context.Context, uploadID string) (*models.Upload, error) {
	if s.uploadRepo == nil {
		return nil, fmt.Errorf("upload service is not initialized")
	}
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return nil, fmt.Errorf("%w: upload_id is required", ErrInvalidUploadRequest)
	}
	return s.uploadRepo.GetUploadByID(ctx, uploadID)
}

func (s *UploadService) maxUploadBytes(ctx context.Context) (int64, error) {
	setting, err := s.settingsRepo.GetSetting(ctx, defaultTenantID, "max_upload_size_mb")
	if err != nil {
		return 0, err
	}

	rawMB, err := parseNumericJSON(setting.SettingValue)
	if err != nil {
		return 0, fmt.Errorf("parse max_upload_size_mb: %w", err)
	}
	if rawMB <= 0 {
		return 0, fmt.Errorf("%w: max_upload_size_mb is not configured", ErrInvalidUploadRequest)
	}
	return int64(rawMB * 1024 * 1024), nil
}

func parseNumericJSON(raw json.RawMessage) (float64, error) {
	var value any
	if len(raw) == 0 {
		return 0, errors.New("empty json")
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, err
	}
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported json numeric type %T", value)
	}
}

func parseUploadMetadata(raw json.RawMessage) (uploadMetadata, error) {
	var metadata uploadMetadata
	if len(raw) == 0 {
		return metadata, errors.New("missing upload metadata")
	}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return metadata, err
	}
	if metadata.FileName == "" {
		return metadata, errors.New("file_name is required in upload metadata")
	}
	return metadata, nil
}

func validateUploadMetadata(metadata uploadMetadata) error {
	if strings.TrimSpace(metadata.FileName) == "" {
		return errors.New("file_name is required")
	}
	return nil
}

func sanitizeFileName(fileName string) (string, error) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", errors.New("file_name cannot be empty")
	}
	if strings.ContainsAny(trimmed, "\x00\r\n") {
		return "", errors.New("file_name contains invalid control characters")
	}
	replaced := strings.ReplaceAll(trimmed, "\\", "/")
	base := path.Base(replaced)
	if base == "." || base == "/" || base == ".." {
		return "", errors.New("file_name is not a valid object name")
	}
	clean := path.Clean(base)
	if clean == "." || clean == "/" {
		return "", errors.New("file_name is not a valid object name")
	}
	if strings.Contains(clean, "/") {
		return "", errors.New("file_name cannot include path separators")
	}
	if strings.Contains(clean, "..") {
		return "", errors.New("file_name cannot include path traversal segments")
	}
	if len(clean) > 255 {
		return "", errors.New("file_name exceeds maximum allowed length")
	}
	return clean, nil
}

func storagePrefix(tenantID string, uploadID string) string {
	if uploadID == "" {
		return path.Join("raw", tenantID)
	}
	return path.Join("raw", tenantID, uploadID)
}

func storageObjectKey(prefix, fileName string) string {
	return path.Join(prefix, fileName)
}

func normalizeExtension(fileName string) string {
	ext := path.Ext(fileName)
	if ext == "" {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(ext), ".")
}

func inferDetectedFileType(fileName, contentType string) string {
	if fileType := normalizeExtension(fileName); fileType != "" {
		return fileType
	}
	return contentType
}

func isArchive(fileName, contentType string) bool {
	lowerName := strings.ToLower(fileName)
	lowerType := strings.ToLower(contentType)

	archiveExtensions := []string{".zip", ".7z", ".rar", ".tar", ".gz", ".tgz", ".tar.gz", ".tar.bz2", ".tar.xz"}
	for _, ext := range archiveExtensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	if strings.Contains(lowerType, "application/zip") ||
		strings.Contains(lowerType, "application/x-7z-compressed") ||
		strings.Contains(lowerType, "application/x-rar") ||
		strings.Contains(lowerType, "application/gzip") {
		return true
	}
	return false
}

func ptrString(value string) *string {
	return &value
}

func rewritePresignedURL(rawURLValue, publicEndpoint string) (string, error) {
	rawURLValue = strings.TrimSpace(rawURLValue)
	publicEndpoint = strings.TrimSpace(publicEndpoint)
	if rawURLValue == "" {
		return "", errors.New("raw presigned url is required")
	}
	if publicEndpoint == "" {
		return rawURLValue, nil
	}

	publicEndpointURL, err := url.Parse(publicEndpoint)
	if err != nil {
		if !strings.Contains(publicEndpoint, "://") {
			publicEndpointURL, err = url.Parse("http://" + publicEndpoint)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	originalURL, err := url.Parse(rawURLValue)
	if err != nil {
		return "", err
	}

	originalURL.Scheme = publicEndpointURL.Scheme
	originalURL.Host = publicEndpointURL.Host
	if publicEndpointURL.Path != "" && publicEndpointURL.Path != "/" {
		originalURL.Path = strings.TrimSuffix(publicEndpointURL.Path, "/") + "/" + strings.TrimPrefix(originalURL.Path, "/")
	}

	return originalURL.String(), nil
}

func ptrFromStringValue(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func ptrFloat64(value float64) *float64 {
	return &value
}
