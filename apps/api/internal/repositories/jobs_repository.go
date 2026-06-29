package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *models.IngestionJob) (*models.IngestionJob, error)
	GetJobByID(ctx context.Context, id string) (*models.IngestionJob, error)
	ListJobs(ctx context.Context, tenantID string) ([]*models.IngestionJob, error)
	ListJobsPaginated(ctx context.Context, tenantID string, limit int, offset int) ([]*models.IngestionJob, error)
	CountJobs(ctx context.Context, tenantID string) (int64, error)
	UpdateJobStatus(ctx context.Context, id string, status string, currentStage *string, progressPercent *float64, errorCode *string, errorMessage *string, startedAt *time.Time, completedAt *time.Time) error
	IncrementRetry(ctx context.Context, id string) error
	GetLatestJobForFile(ctx context.Context, tenantID string, fileID string) (*models.IngestionJob, error)
}

type JobStatusUpdate struct {
	ID              string
	Status          string
	CurrentStage    *string
	ProgressPercent *float64
	ErrorCode       *string
	ErrorMessage    *string
	StartedAt       *time.Time
	CompletedAt     *time.Time
}

type PostgresJobRepository struct {
	pool *pgxpool.Pool
}

func NewJobRepository(pool *pgxpool.Pool) *PostgresJobRepository {
	return &PostgresJobRepository{pool: pool}
}

func (r *PostgresJobRepository) CreateJob(ctx context.Context, job *models.IngestionJob) (*models.IngestionJob, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job is required")
	}
	if job.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if job.RootFileID == "" {
		return nil, fmt.Errorf("root_file_id is required")
	}
	if job.ID == "" {
		job.ID = uuid.NewString()
	}
	if job.Status == "" {
		job.Status = "queued"
	}

	query := `
		INSERT INTO ingestion_jobs (
			id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count,
			error_code, error_message, started_at, completed_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
		RETURNING id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count, error_code, error_message, started_at, completed_at, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		job.ID,
		job.TenantID,
		job.RootFileID,
		job.Status,
		nullString(job.CurrentStage),
		job.ProgressPercent,
		job.RetryCount,
		nullString(job.ErrorCode),
		nullString(job.ErrorMessage),
		job.StartedAt,
		job.CompletedAt,
	)
	return scanJob(row)
}

func (r *PostgresJobRepository) GetJobByID(ctx context.Context, id string) (*models.IngestionJob, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("job id is required")
	}
	query := `
		SELECT id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count, error_code, error_message, started_at, completed_at, created_at, updated_at
		FROM ingestion_jobs
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	job, err := scanJob(row)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return job, nil
}

func (r *PostgresJobRepository) ListJobs(ctx context.Context, tenantID string) ([]*models.IngestionJob, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	query := `
		SELECT id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count, error_code, error_message, started_at, completed_at, created_at, updated_at
		FROM ingestion_jobs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.IngestionJob, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresJobRepository) ListJobsPaginated(ctx context.Context, tenantID string, limit int, offset int) ([]*models.IngestionJob, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count, error_code, error_message, started_at, completed_at, created_at, updated_at
		FROM ingestion_jobs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.IngestionJob, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresJobRepository) CountJobs(ctx context.Context, tenantID string) (int64, error) {
	if err := validatePool(r.pool); err != nil {
		return 0, err
	}
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ingestion_jobs WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PostgresJobRepository) UpdateJobStatus(
	ctx context.Context,
	id string,
	status string,
	currentStage *string,
	progressPercent *float64,
	errorCode *string,
	errorMessage *string,
	startedAt *time.Time,
	completedAt *time.Time,
) error {
	if err := validatePool(r.pool); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("job id is required")
	}
	if status == "" {
		return fmt.Errorf("status is required")
	}

	query := `
		UPDATE ingestion_jobs
		SET status = $2,
		    current_stage = COALESCE($3, current_stage),
		    progress_percent = COALESCE($4, progress_percent),
		    error_code = COALESCE($5, error_code),
		    error_message = COALESCE($6, error_message),
		    started_at = COALESCE($7, started_at),
		    completed_at = COALESCE($8, completed_at),
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, status, currentStage, progressPercent, errorCode, errorMessage, startedAt, completedAt)
	if err != nil {
		return err
	}
	return rowsAffectedError(tag, "job")
}

func (r *PostgresJobRepository) IncrementRetry(ctx context.Context, id string) error {
	if err := validatePool(r.pool); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("job id is required")
	}
	query := `
		UPDATE ingestion_jobs
		SET retry_count = retry_count + 1,
			updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return rowsAffectedError(tag, "job")
}

func (r *PostgresJobRepository) GetLatestJobForFile(ctx context.Context, tenantID string, fileID string) (*models.IngestionJob, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	tenantID = strings.TrimSpace(tenantID)
	fileID = strings.TrimSpace(fileID)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if fileID == "" {
		return nil, fmt.Errorf("file_id is required")
	}

	query := `
		WITH RECURSIVE file_tree AS (
			SELECT id, parent_file_id
			FROM files
			WHERE tenant_id = $1 AND id = $2
			UNION ALL
			SELECT parent.id, parent.parent_file_id
			FROM files AS parent
			JOIN file_tree AS child ON parent.id = child.parent_file_id
			WHERE parent.tenant_id = $1
		)
		SELECT
			id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count,
			error_code, error_message, started_at, completed_at, created_at, updated_at
		FROM ingestion_jobs
		WHERE tenant_id = $1 AND root_file_id IN (SELECT id FROM file_tree)
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, tenantID, fileID)
	job, err := scanJob(row)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func scanJob(scanner rowScanner) (*models.IngestionJob, error) {
	var (
		id, tenantID, rootFileID                      string
		status                                        string
		currentStage, errorCode, errorMessage          sql.NullString
		progressPercent                               float64
		retryCount                                    int
		startedAt, completedAt                        sql.NullTime
		createdAt, updatedAt                          sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&tenantID,
		&rootFileID,
		&status,
		&currentStage,
		&progressPercent,
		&retryCount,
		&errorCode,
		&errorMessage,
		&startedAt,
		&completedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, mapNoRows(err)
	}

	ca := createdAt.Time
	ua := updatedAt.Time
	if !createdAt.Valid {
		ca = time.Time{}
	}
	if !updatedAt.Valid {
		ua = time.Time{}
	}

	return &models.IngestionJob{
		ID:              id,
		TenantID:        tenantID,
		RootFileID:      rootFileID,
		Status:          status,
		CurrentStage:    nullPtrFromString(currentStage),
		ProgressPercent: progressPercent,
		RetryCount:      retryCount,
		ErrorCode:       nullPtrFromString(errorCode),
		ErrorMessage:    nullPtrFromString(errorMessage),
		StartedAt:       nullTimePtr(startedAt),
		CompletedAt:     nullTimePtr(completedAt),
		CreatedAt:       ca,
		UpdatedAt:       ua,
	}, nil
}

func nullTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	value := v.Time
	return &value
}
