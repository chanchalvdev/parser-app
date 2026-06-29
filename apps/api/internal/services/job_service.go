package services

import (
	"context"
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
	defaultJobListPageSize = 25
	maxJobListPageSize     = 200
)

var (
	ErrInvalidJobRequest = errors.New("invalid job request")
)

type JobService struct {
	jobRepo       repositories.JobRepository
	jobEventRepo  repositories.JobEventRepository
	fileRepo      repositories.FileRepository
	auditLogRepo  repositories.AuditLogRepository
	queueProducer queue.Producer
	minioBucket   string
}

type ListJobsRequest struct {
	TenantID string
	Page     int
	PageSize int
}

type JobListResponse struct {
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Jobs     []*models.IngestionJob `json:"jobs"`
}

type ListJobEventsRequest struct {
	JobID    string
	Page     int
	PageSize int
}

type JobEventsResponse struct {
	JobID    string             `json:"job_id"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Events   []*models.JobEvent `json:"events"`
}

type RetryJobRequest struct {
	IPAddress *string
	UserAgent *string
}

func NewJobService(
	jobRepo repositories.JobRepository,
	jobEventRepo repositories.JobEventRepository,
	fileRepo repositories.FileRepository,
	auditLogRepo repositories.AuditLogRepository,
	queueProducer queue.Producer,
	minioBucket string,
) *JobService {
	return &JobService{
		jobRepo:       jobRepo,
		jobEventRepo:  jobEventRepo,
		fileRepo:      fileRepo,
		auditLogRepo:  auditLogRepo,
		queueProducer: queueProducer,
		minioBucket:   minioBucket,
	}
}

func (s *JobService) ListJobs(ctx context.Context, req ListJobsRequest) (*JobListResponse, error) {
	if s.jobRepo == nil {
		return nil, fmt.Errorf("job service is not initialized: job repository")
	}

	req = normalizeJobListRequest(req)
	if req.TenantID == "" {
		req.TenantID = defaultTenantID
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	jobs, err := s.jobRepo.ListJobsPaginated(ctx, req.TenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.jobRepo.CountJobs(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	return &JobListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Jobs:     jobs,
	}, nil
}

func (s *JobService) GetJob(ctx context.Context, jobID string) (*models.IngestionJob, error) {
	if s.jobRepo == nil {
		return nil, fmt.Errorf("job service is not initialized: job repository")
	}

	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("%w: job_id is required", ErrInvalidJobRequest)
	}
	return s.jobRepo.GetJobByID(ctx, jobID)
}

func (s *JobService) ListJobEvents(ctx context.Context, req ListJobEventsRequest) (*JobEventsResponse, error) {
	if s.jobRepo == nil || s.jobEventRepo == nil {
		return nil, fmt.Errorf("job service is not initialized")
	}

	req = normalizeJobEventsRequest(req)
	if req.JobID == "" {
		return nil, fmt.Errorf("%w: job_id is required", ErrInvalidJobRequest)
	}

	if _, err := s.jobRepo.GetJobByID(ctx, req.JobID); err != nil {
		return nil, err
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize
	events, err := s.jobEventRepo.ListJobEventsPaginated(ctx, req.JobID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.jobEventRepo.CountJobEvents(ctx, req.JobID)
	if err != nil {
		return nil, err
	}

	return &JobEventsResponse{
		JobID:    req.JobID,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Events:   events,
	}, nil
}

func (s *JobService) RetryJob(ctx context.Context, jobID string, req RetryJobRequest) (*models.IngestionJob, error) {
	if s.jobRepo == nil || s.jobEventRepo == nil || s.fileRepo == nil || s.auditLogRepo == nil || s.queueProducer == nil {
		return nil, fmt.Errorf("job service is not fully initialized")
	}

	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("%w: job_id is required", ErrInvalidJobRequest)
	}
	job, err := s.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	rootFile, err := s.fileRepo.GetFileByID(ctx, job.RootFileID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	queueMessage := queue.IngestionJobMessage{
		JobID:       job.ID,
		TenantID:    job.TenantID,
		RootFileID:  job.RootFileID,
		StoragePath: rootFile.StoragePath,
		Bucket:      s.minioBucket,
		OriginalName: rootFile.OriginalName,
		Depth:       rootFile.Depth,
		CreatedAt:   now.Format(time.RFC3339),
	}
	if err := s.queueProducer.EnqueueIngestionJob(ctx, queueMessage); err != nil {
		if eventErr := s.createRetryEvent(
			ctx,
			job,
			"job.retry.failed",
			fmt.Sprintf("failed to enqueue job for retry: %v", err),
			map[string]any{
				"root_file_id": job.RootFileID,
				"file_id":      job.RootFileID,
				"status":       "failed",
				"error":        err.Error(),
				"occurred_at":  now.Format(time.RFC3339),
			},
		); eventErr != nil {
			return nil, eventErr
		}
		if auditErr := s.createRetryAuditLog(ctx, job, req, "failed", err.Error(), now); auditErr != nil {
			return nil, auditErr
		}
		return nil, fmt.Errorf("enqueue retry job: %w", err)
	}

	if err := s.jobRepo.IncrementRetry(ctx, jobID); err != nil {
		return nil, err
	}

	updatedJob, err := s.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if err := s.createRetryEvent(
		ctx,
		updatedJob,
		"job.retry.completed",
		"job requeued for retry",
		map[string]any{
			"root_file_id": job.RootFileID,
			"retry_count":  updatedJob.RetryCount,
			"status":       "queued",
			"requeued_at":  now.Format(time.RFC3339),
		},
	); err != nil {
		return nil, err
	}

	if err := s.createRetryAuditLog(ctx, updatedJob, req, "queued", "job requeued for retry", now); err != nil {
		return nil, err
	}

	return updatedJob, nil
}

func (s *JobService) createRetryEvent(
	ctx context.Context,
	job *models.IngestionJob,
	eventType string,
	eventMessage string,
	details map[string]any,
) error {
	eventDetails, err := json.Marshal(details)
	if err != nil {
		return err
	}
	evtMessage := eventMessage
	_, err = s.jobEventRepo.CreateJobEvent(ctx, &models.JobEvent{
		TenantID:     job.TenantID,
		JobID:        job.ID,
		EventType:    eventType,
		EventMessage: &evtMessage,
		EventDetails: eventDetails,
		CreatedBy:    ptrString(defaultUserID),
	})
	return err
}

func (s *JobService) createRetryAuditLog(
	ctx context.Context,
	job *models.IngestionJob,
	req RetryJobRequest,
	result string,
	message string,
	occurredAt time.Time,
) error {
	details := map[string]any{
		"job_id":      job.ID,
		"result":      result,
		"message":     message,
		"retry_count": job.RetryCount,
		"occurred_at": occurredAt.Format(time.RFC3339),
	}
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = s.auditLogRepo.CreateAuditLog(ctx, &models.AuditLog{
		TenantID:   job.TenantID,
		ActorUserID: ptrString(defaultUserID),
		Action:     "job.retried",
		EntityType: ptrString("job"),
		EntityID:   ptrString(job.ID),
		Details:    detailsJSON,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	})
	return err
}

func normalizeJobListRequest(req ListJobsRequest) ListJobsRequest {
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.Page = normalizePage(req.Page, 1)
	req.PageSize = normalizePageSize(req.PageSize, defaultJobListPageSize, maxJobListPageSize)
	return req
}

func normalizeJobEventsRequest(req ListJobEventsRequest) ListJobEventsRequest {
	req.JobID = strings.TrimSpace(req.JobID)
	req.Page = normalizePage(req.Page, 1)
	req.PageSize = normalizePageSize(req.PageSize, defaultJobListPageSize, maxJobListPageSize)
	return req
}
