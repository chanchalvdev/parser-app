package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobEventRepository interface {
	CreateJobEvent(ctx context.Context, event *models.JobEvent) (*models.JobEvent, error)
	ListJobEvents(ctx context.Context, jobID string) ([]*models.JobEvent, error)
	ListJobEventsPaginated(ctx context.Context, jobID string, limit int, offset int) ([]*models.JobEvent, error)
	CountJobEvents(ctx context.Context, jobID string) (int64, error)
}

type PostgresJobEventRepository struct {
	pool *pgxpool.Pool
}

func NewJobEventRepository(pool *pgxpool.Pool) *PostgresJobEventRepository {
	return &PostgresJobEventRepository{pool: pool}
}

func (r *PostgresJobEventRepository) CreateJobEvent(ctx context.Context, event *models.JobEvent) (*models.JobEvent, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event is required")
	}
	if event.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if event.JobID == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}
	if event.EventMessage == nil {
		return nil, fmt.Errorf("event_message is required")
	}

	details, err := asJSON(event.EventDetails)
	if err != nil {
		return nil, fmt.Errorf("event details: %w", err)
	}

	query := `
		INSERT INTO job_events (
			id, tenant_id, job_id, event_type, event_message, event_details, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, tenant_id, job_id, event_type, event_message, event_details, created_by, created_at
	`
	row := r.pool.QueryRow(
		ctx,
		query,
		event.ID,
		event.TenantID,
		event.JobID,
		event.EventType,
		event.EventMessage,
		details,
		nullString(event.CreatedBy),
	)
	return scanJobEvent(row)
}

func (r *PostgresJobEventRepository) ListJobEvents(ctx context.Context, jobID string) ([]*models.JobEvent, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if jobID == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	query := `
		SELECT id, tenant_id, job_id, event_type, event_message, event_details, created_by, created_at
		FROM job_events
		WHERE job_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.JobEvent, 0)
	for rows.Next() {
		event, err := scanJobEvent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresJobEventRepository) ListJobEventsPaginated(ctx context.Context, jobID string, limit int, offset int) ([]*models.JobEvent, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if jobID == "" {
		return nil, fmt.Errorf("job_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, job_id, event_type, event_message, event_details, created_by, created_at
		FROM job_events
		WHERE job_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, jobID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.JobEvent, 0)
	for rows.Next() {
		event, err := scanJobEvent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresJobEventRepository) CountJobEvents(ctx context.Context, jobID string) (int64, error) {
	if err := validatePool(r.pool); err != nil {
		return 0, err
	}
	if jobID == "" {
		return 0, fmt.Errorf("job_id is required")
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM job_events WHERE job_id = $1`, jobID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func scanJobEvent(scanner rowScanner) (*models.JobEvent, error) {
	var (
		id, tenantID, jobID, eventType, eventMessage sql.NullString
		eventDetails                               []byte
		createdBy                                  sql.NullString
		createdAt                                  sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&tenantID,
		&jobID,
		&eventType,
		&eventMessage,
		&eventDetails,
		&createdBy,
		&createdAt,
	); err != nil {
		return nil, mapNoRows(err)
	}
	if !id.Valid || !tenantID.Valid || !jobID.Valid || !eventType.Valid {
		return nil, fmt.Errorf("incomplete job event row")
	}

	details, err := scanJSON(eventDetails)
	if err != nil {
		return nil, err
	}
	createdAtTime := createdAt.Time
	if !createdAt.Valid {
		createdAtTime = time.Time{}
	}

	return &models.JobEvent{
		ID:           id.String,
		TenantID:     tenantID.String,
		JobID:        jobID.String,
		EventType:    eventType.String,
		EventMessage: nullPtrFromString(eventMessage),
		EventDetails: details,
		CreatedBy:    nullPtrFromString(createdBy),
		CreatedAt:    createdAtTime,
	}, nil
}
