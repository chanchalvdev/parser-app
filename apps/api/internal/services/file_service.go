package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/queue"
	"github.com/enterprise-file-platform/api/internal/repositories"
)

const (
	defaultFileListPageSize         = 25
	maxFileListPageSize             = 200
	passwordRefAlgorithmLocalBase64  = "local_base64"
	archivePasswordStatusRequired   = "password_required"
	archivePasswordStatusWrong      = "wrong_password"
	archivePasswordSubmittedAction   = "password.submitted"
)

var (
	ErrInvalidFileRequest         = errors.New("invalid file request")
	ErrInvalidFilePasswordRequest = errors.New("invalid file password request")
)

type FileService struct {
	fileRepo              repositories.FileRepository
	parsedRecordRepo       repositories.ParsedRecordRepository
	jobRepo               repositories.JobRepository
	jobEventRepo          repositories.JobEventRepository
	archivePasswordRefRepo repositories.ArchivePasswordRefRepository
	auditLogRepo          repositories.AuditLogRepository
	queueProducer         queue.Producer
	minioBucket           string
}

type ListFilesRequest struct {
	TenantID         string
	ProcessingStatus string
	Extension        string
	DetectedFileType string
	Page             int
	PageSize         int
}

type FileListResponse struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Files    []*models.File `json:"files"`
}

type ListFileChildrenRequest struct {
	FileID   string
	Page     int
	PageSize int
}

type FileChildrenResponse struct {
	FileID   string         `json:"file_id"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Children []*models.File `json:"children"`
}

type ListFileRecordsRequest struct {
	FileID   string
	Page     int
	PageSize int
}

type FileRecordsResponse struct {
	FileID   string                 `json:"file_id"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Records  []*models.ParsedRecord `json:"records"`
}

type FileTreeNode struct {
	File     *models.File    `json:"file"`
	Children []*FileTreeNode `json:"children"`
}

type SubmitFilePasswordRequest struct {
	Password string `json:"password"`
}

type SubmitFilePasswordResponse struct {
	FileID string `json:"file_id"`
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type SubmitFilePasswordContext struct {
	IPAddress *string
	UserAgent *string
}

func NewFileService(
	fileRepo repositories.FileRepository,
	parsedRecordRepo repositories.ParsedRecordRepository,
	jobRepo repositories.JobRepository,
	jobEventRepo repositories.JobEventRepository,
	archivePasswordRefRepo repositories.ArchivePasswordRefRepository,
	auditLogRepo repositories.AuditLogRepository,
	queueProducer queue.Producer,
	minioBucket string,
) *FileService {
	return &FileService{
		fileRepo:              fileRepo,
		parsedRecordRepo:       parsedRecordRepo,
		jobRepo:               jobRepo,
		jobEventRepo:          jobEventRepo,
		archivePasswordRefRepo: archivePasswordRefRepo,
		auditLogRepo:          auditLogRepo,
		queueProducer:         queueProducer,
		minioBucket:           minioBucket,
	}
}

func (s *FileService) ListFiles(ctx context.Context, req ListFilesRequest) (*FileListResponse, error) {
	if s.fileRepo == nil {
		return nil, fmt.Errorf("file service is not initialized: file repository")
	}

	req = normalizeFileListRequest(req)
	if req.TenantID == "" {
		req.TenantID = defaultTenantID
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	files, err := s.fileRepo.ListFilesFiltered(ctx, req.TenantID, req.ProcessingStatus, req.Extension, req.DetectedFileType, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.fileRepo.CountFiles(ctx, req.TenantID, req.ProcessingStatus, req.Extension, req.DetectedFileType)
	if err != nil {
		return nil, err
	}

	return &FileListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Files:    files,
	}, nil
}

func (s *FileService) GetFile(ctx context.Context, fileID string) (*models.File, error) {
	if s.fileRepo == nil {
		return nil, fmt.Errorf("file service is not initialized: file repository")
	}

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", ErrInvalidFileRequest)
	}
	return s.fileRepo.GetFileByID(ctx, fileID)
}

func (s *FileService) ListFileChildren(ctx context.Context, req ListFileChildrenRequest) (*FileChildrenResponse, error) {
	if s.fileRepo == nil {
		return nil, fmt.Errorf("file service is not initialized: file repository")
	}

	req = normalizeFileChildrenRequest(req)
	if req.FileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", ErrInvalidFileRequest)
	}

	_, err := s.fileRepo.GetFileByID(ctx, req.FileID)
	if err != nil {
		return nil, err
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	children, err := s.fileRepo.ListChildrenPaginated(ctx, req.FileID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.fileRepo.CountChildren(ctx, req.FileID)
	if err != nil {
		return nil, err
	}

	return &FileChildrenResponse{
		FileID:   req.FileID,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Children: children,
	}, nil
}

func (s *FileService) GetFileTree(ctx context.Context, fileID string) (*FileTreeNode, error) {
	if s.fileRepo == nil {
		return nil, fmt.Errorf("file service is not initialized: file repository")
	}

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", ErrInvalidFileRequest)
	}

	root, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	tree := &FileTreeNode{
		File:     root,
		Children: make([]*FileTreeNode, 0),
	}

	visited := map[string]struct{}{
		root.ID: {},
	}
	if err := s.buildFileTree(ctx, tree, visited); err != nil {
		return nil, err
	}

	return tree, nil
}

func (s *FileService) ListFileRecords(ctx context.Context, req ListFileRecordsRequest) (*FileRecordsResponse, error) {
	if s.fileRepo == nil || s.parsedRecordRepo == nil {
		return nil, fmt.Errorf("file service is not initialized")
	}

	req = normalizeFileRecordsRequest(req)
	if req.FileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", ErrInvalidFileRequest)
	}

	if _, err := s.fileRepo.GetFileByID(ctx, req.FileID); err != nil {
		return nil, err
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	records, err := s.parsedRecordRepo.ListRecordsByFileID(ctx, req.FileID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.parsedRecordRepo.CountRecordsByFileID(ctx, req.FileID)
	if err != nil {
		return nil, err
	}

	return &FileRecordsResponse{
		FileID:   req.FileID,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Records:  records,
	}, nil
}

func (s *FileService) SubmitFilePassword(
	ctx context.Context,
	fileID string,
	req SubmitFilePasswordRequest,
	context SubmitFilePasswordContext,
) (*SubmitFilePasswordResponse, error) {
	if s.fileRepo == nil || s.jobRepo == nil || s.jobEventRepo == nil || s.archivePasswordRefRepo == nil || s.auditLogRepo == nil || s.queueProducer == nil {
		return nil, fmt.Errorf("file service is not fully initialized")
	}

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", ErrInvalidFilePasswordRequest)
	}

	password := strings.TrimSpace(req.Password)
	if password == "" {
		return nil, fmt.Errorf("%w: password is required", ErrInvalidFilePasswordRequest)
	}

	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	status := strings.ToLower(file.ProcessingStatus)
	if status != archivePasswordStatusRequired && status != archivePasswordStatusWrong {
		return nil, fmt.Errorf("%w: file is not waiting for password", ErrInvalidFilePasswordRequest)
	}

	passwordRefHash, err := encodePasswordRef(password)
	if err != nil {
		return nil, err
	}

	ref := &models.ArchivePasswordRef{
		TenantID:       file.TenantID,
		FileID:         file.ID,
		UploadID:       file.UploadID,
		PasswordRefHash: passwordRefHash,
		Algorithm:      passwordRefAlgorithmLocalBase64,
		IsValid:        true,
		Validated:      false,
		AttemptCount:   0,
		CreatedBy:      ptrString(defaultUserID),
	}
	if _, err := s.archivePasswordRefRepo.CreateArchivePasswordRef(ctx, ref); err != nil {
		return nil, err
	}

	job, err := s.jobRepo.GetLatestJobForFile(ctx, file.TenantID, file.ID)
	if err != nil {
		return nil, err
	}

	rootFile, err := s.fileRepo.GetFileByID(ctx, job.RootFileID)
	if err != nil {
		return nil, err
	}

	if err := s.requeueJobWithPassword(ctx, file, rootFile, job); err != nil {
		return nil, err
	}

	if err := s.createPasswordSubmittedEvent(ctx, job, file.ID, "password submitted and job requeued"); err != nil {
		return nil, err
	}
	if err := s.createPasswordSubmittedAuditLog(ctx, job, file.ID, archivePasswordSubmittedAction, context); err != nil {
		return nil, err
	}

	if err := s.jobRepo.UpdateJobStatus(
		ctx,
		job.ID,
		"queued",
		ptrString("queued"),
		ptrFloat64(0),
		nil,
		nil,
		nil,
		nil,
	); err != nil {
		// keep job submitted but allow UI to reflect explicit event if status update fails
		return nil, err
	}

	if err := s.fileRepo.UpdateFileStatus(ctx, file.ID, "queued"); err != nil {
		return nil, err
	}

	return &SubmitFilePasswordResponse{
		FileID: file.ID,
		JobID:  job.ID,
		Status: "queued",
	}, nil
}

func (s *FileService) requeueJobWithPassword(
	ctx context.Context,
	file *models.File,
	rootFile *models.File,
	job *models.IngestionJob,
) error {
	queueMessage := queue.IngestionJobMessage{
		JobID:       job.ID,
		TenantID:    file.TenantID,
		RootFileID:  rootFile.ID,
		StoragePath: rootFile.StoragePath,
		Bucket:      s.minioBucket,
		OriginalName: rootFile.OriginalName,
		Depth:       rootFile.Depth,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.queueProducer.EnqueueIngestionJob(ctx, queueMessage); err != nil {
		return err
	}

	if err := s.jobRepo.IncrementRetry(ctx, job.ID); err != nil {
		return err
	}
	return nil
}

func (s *FileService) createPasswordSubmittedEvent(ctx context.Context, job *models.IngestionJob, fileID string, message string) error {
	eventDetails, err := json.Marshal(map[string]any{
		"file_id":   fileID,
		"job_id":    job.ID,
		"root_file": job.RootFileID,
		"status":    "queued",
		"message":   message,
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	msg := message
	_, err = s.jobEventRepo.CreateJobEvent(ctx, &models.JobEvent{
		TenantID:     job.TenantID,
		JobID:        job.ID,
		EventType:    "worker.password_submitted",
		EventMessage: &msg,
		EventDetails: eventDetails,
		CreatedBy:    ptrString(defaultUserID),
	})
	return err
}

func (s *FileService) createPasswordSubmittedAuditLog(
	ctx context.Context,
	job *models.IngestionJob,
	fileID string,
	action string,
	requestCtx SubmitFilePasswordContext,
) error {
	details, err := json.Marshal(map[string]any{
		"file_id":      fileID,
		"job_id":       job.ID,
		"root_file_id": job.RootFileID,
		"status":       "queued",
		"occurred_at":  time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	_, err = s.auditLogRepo.CreateAuditLog(ctx, &models.AuditLog{
		TenantID:    job.TenantID,
		ActorUserID: ptrString(defaultUserID),
		Action:      action,
		EntityType:  ptrString("file"),
		EntityID:    ptrString(fileID),
		Details:     details,
		IPAddress:   requestCtx.IPAddress,
		UserAgent:   requestCtx.UserAgent,
	})
	return err
}

func encodePasswordRef(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("%w: password is required", ErrInvalidFilePasswordRequest)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(password))
	return encoded, nil
}

func (s *FileService) buildFileTree(ctx context.Context, node *FileTreeNode, visited map[string]struct{}) error {
	children, err := s.fileRepo.ListChildren(ctx, node.File.ID)
	if err != nil {
		return err
	}
	for _, child := range children {
		if _, exists := visited[child.ID]; exists {
			continue
		}
		visited[child.ID] = struct{}{}
		childNode := &FileTreeNode{
			File:     child,
			Children: make([]*FileTreeNode, 0),
		}
		if err := s.buildFileTree(ctx, childNode, visited); err != nil {
			return err
		}
		node.Children = append(node.Children, childNode)
	}
	return nil
}

func normalizeFileListRequest(req ListFilesRequest) ListFilesRequest {
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.ProcessingStatus = strings.TrimSpace(req.ProcessingStatus)
	req.Extension = strings.TrimSpace(req.Extension)
	req.DetectedFileType = strings.TrimSpace(req.DetectedFileType)
	req.Page = normalizePage(req.Page, 1)
	req.PageSize = normalizePageSize(req.PageSize, defaultFileListPageSize, maxFileListPageSize)
	return req
}

func normalizeFileChildrenRequest(req ListFileChildrenRequest) ListFileChildrenRequest {
	req.FileID = strings.TrimSpace(req.FileID)
	req.Page = normalizePage(req.Page, 1)
	req.PageSize = normalizePageSize(req.PageSize, defaultFileListPageSize, maxFileListPageSize)
	return req
}

func normalizeFileRecordsRequest(req ListFileRecordsRequest) ListFileRecordsRequest {
	req.FileID = strings.TrimSpace(req.FileID)
	req.Page = normalizePage(req.Page, 1)
	req.PageSize = normalizePageSize(req.PageSize, defaultFileListPageSize, maxFileListPageSize)
	return req
}

func normalizePage(page int, defaultPage int) int {
	if page <= 0 {
		return defaultPage
	}
	return page
}

func normalizePageSize(pageSize int, defaultSize int, maxSize int) int {
	if pageSize <= 0 {
		return defaultSize
	}
	if pageSize > maxSize {
		return maxSize
	}
	return pageSize
}
