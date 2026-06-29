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

type ArchivePasswordRefRepository interface {
	CreateArchivePasswordRef(ctx context.Context, ref *models.ArchivePasswordRef) (*models.ArchivePasswordRef, error)
	GetLatestArchivePasswordRef(ctx context.Context, tenantID string, fileID string) (*models.ArchivePasswordRef, error)
}

type PostgresArchivePasswordRefRepository struct {
	pool *pgxpool.Pool
}

func NewArchivePasswordRefRepository(pool *pgxpool.Pool) *PostgresArchivePasswordRefRepository {
	return &PostgresArchivePasswordRefRepository{pool: pool}
}

func (r *PostgresArchivePasswordRefRepository) CreateArchivePasswordRef(
	ctx context.Context,
	ref *models.ArchivePasswordRef,
) (*models.ArchivePasswordRef, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, fmt.Errorf("archive password ref is required")
	}
	if ref.TenantID == "" || ref.FileID == "" {
		return nil, fmt.Errorf("tenant_id and file_id are required")
	}
	if ref.PasswordRefHash == "" {
		return nil, fmt.Errorf("password_ref_hash is required")
	}
	if ref.Algorithm == "" {
		return nil, fmt.Errorf("algorithm is required")
	}
	if ref.ID == "" {
		ref.ID = uuid.NewString()
	}

	query := `
		INSERT INTO archive_password_refs (
			id, tenant_id, file_id, upload_id, password_ref_hash, algorithm, is_valid,
			validated, attempt_count, last_validated_at, created_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, file_id, password_ref_hash) DO UPDATE
		SET is_valid = EXCLUDED.is_valid,
		    validated = EXCLUDED.validated,
		    attempt_count = EXCLUDED.attempt_count,
		    last_validated_at = EXCLUDED.last_validated_at,
		    created_by = EXCLUDED.created_by,
		    upload_id = EXCLUDED.upload_id,
		    updated_at = NOW()
		RETURNING
			id, tenant_id, file_id, upload_id, password_ref_hash, algorithm,
			is_valid, validated, attempt_count, last_validated_at, created_by, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		ref.ID,
		ref.TenantID,
		ref.FileID,
		nullString(ref.UploadID),
		ref.PasswordRefHash,
		ref.Algorithm,
		ref.IsValid,
		ref.Validated,
		ref.AttemptCount,
		ref.LastValidated,
		nullString(ref.CreatedBy),
	)

	var (
		uploadID         sql.NullString
		createdBy        sql.NullString
		lastValidatedAt  sql.NullTime
		createdAt        time.Time
		updatedAt        time.Time
	)

	var response models.ArchivePasswordRef
	if err := row.Scan(
		&response.ID,
		&response.TenantID,
		&response.FileID,
		&uploadID,
		&response.PasswordRefHash,
		&response.Algorithm,
		&response.IsValid,
		&response.Validated,
		&response.AttemptCount,
		&lastValidatedAt,
		&createdBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, mapNoRows(err)
	}

	response.UploadID = nullPtrFromString(uploadID)
	response.CreatedBy = nullPtrFromString(createdBy)
	if lastValidatedAt.Valid {
		tm := lastValidatedAt.Time
		response.LastValidated = &tm
	} else {
		response.LastValidated = nil
	}
	response.CreatedAt = createdAt
	response.UpdatedAt = updatedAt

	return &response, nil
}

func (r *PostgresArchivePasswordRefRepository) GetLatestArchivePasswordRef(
	ctx context.Context,
	tenantID string,
	fileID string,
) (*models.ArchivePasswordRef, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	tenantID = strings.TrimSpace(tenantID)
	fileID = strings.TrimSpace(fileID)
	if tenantID == "" || fileID == "" {
		return nil, fmt.Errorf("tenant_id and file_id are required")
	}

	query := `
		SELECT
			id, tenant_id, file_id, upload_id, password_ref_hash, algorithm,
			is_valid, validated, attempt_count, last_validated_at, created_by, created_at, updated_at
		FROM archive_password_refs
		WHERE tenant_id = $1 AND file_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, tenantID, fileID)
	var (
		uploadID        sql.NullString
		createdBy       sql.NullString
		lastValidatedAt sql.NullTime
		createdAt       time.Time
		updatedAt       time.Time
	)

	var response models.ArchivePasswordRef
	if err := row.Scan(
		&response.ID,
		&response.TenantID,
		&response.FileID,
		&uploadID,
		&response.PasswordRefHash,
		&response.Algorithm,
		&response.IsValid,
		&response.Validated,
		&response.AttemptCount,
		&lastValidatedAt,
		&createdBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		if mapNoRows(err) == ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	response.UploadID = nullPtrFromString(uploadID)
	response.CreatedBy = nullPtrFromString(createdBy)
	if lastValidatedAt.Valid {
		tm := lastValidatedAt.Time
		response.LastValidated = &tm
	} else {
		response.LastValidated = nil
	}
	response.CreatedAt = createdAt
	response.UpdatedAt = updatedAt

	return &response, nil
}
