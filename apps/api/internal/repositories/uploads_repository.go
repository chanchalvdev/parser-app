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

type UploadRepository interface {
	CreateUpload(ctx context.Context, upload *models.Upload) (*models.Upload, error)
	GetUploadByID(ctx context.Context, id string) (*models.Upload, error)
	MarkUploadCompleted(ctx context.Context, id string) error
}

type PostgresUploadRepository struct {
	pool *pgxpool.Pool
}

func NewUploadRepository(pool *pgxpool.Pool) *PostgresUploadRepository {
	return &PostgresUploadRepository{pool: pool}
}

func (r *PostgresUploadRepository) CreateUpload(ctx context.Context, upload *models.Upload) (*models.Upload, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if upload == nil {
		return nil, fmt.Errorf("upload is required")
	}
	if upload.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if upload.ID == "" {
		upload.ID = uuid.NewString()
	}
	if upload.Status == "" {
		upload.Status = "pending"
	}

	metadata, err := asJSON(upload.Metadata)
	if err != nil {
		return nil, fmt.Errorf("upload.Metadata: %w", err)
	}

	query := `
		INSERT INTO uploads (
			id, tenant_id, initiated_by, status, original_source, storage_prefix,
			total_files, total_size_bytes, metadata
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING
			id, tenant_id, initiated_by, status, original_source, storage_prefix,
			total_files, total_size_bytes, metadata, created_at, updated_at
	`
	row := r.pool.QueryRow(
		ctx,
		query,
		upload.ID,
		upload.TenantID,
		nullString(upload.InitiatedBy),
		upload.Status,
		nullString(upload.OriginalSource),
		nullString(upload.StoragePrefix),
		upload.TotalFiles,
		upload.TotalSizeBytes,
		metadata,
	)

	return scanUpload(row)
}

func (r *PostgresUploadRepository) GetUploadByID(ctx context.Context, id string) (*models.Upload, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("upload id is required")
	}

	query := `
		SELECT id, tenant_id, initiated_by, status, original_source, storage_prefix,
			total_files, total_size_bytes, metadata, created_at, updated_at
		FROM uploads
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	upload, err := scanUpload(row)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return upload, nil
}

func (r *PostgresUploadRepository) MarkUploadCompleted(ctx context.Context, id string) error {
	if err := validatePool(r.pool); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("upload id is required")
	}
	query := `
		UPDATE uploads
		SET status='completed',
			updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return rowsAffectedError(tag, "upload")
}

func scanUpload(scanner rowScanner) (*models.Upload, error) {
	var (
		id, tenantID         string
		initiatedBy           sql.NullString
		status               string
		originalSource       sql.NullString
		storagePrefix        sql.NullString
		totalFiles           int
		totalSizeBytes       int64
		metadataRaw          []byte
		createdAt, updatedAt  sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&tenantID,
		&initiatedBy,
		&status,
		&originalSource,
		&storagePrefix,
		&totalFiles,
		&totalSizeBytes,
		&metadataRaw,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, mapNoRows(err)
	}

	metadata, err := scanJSON(metadataRaw)
	if err != nil {
		return nil, err
	}
	createdAtTime := createdAt.Time
	updatedAtTime := updatedAt.Time
	if !createdAt.Valid {
		createdAtTime = time.Time{}
	}
	if !updatedAt.Valid {
		updatedAtTime = time.Time{}
	}

	return &models.Upload{
		ID:             id,
		TenantID:       tenantID,
		InitiatedBy:    nullPtrFromString(initiatedBy),
		Status:         status,
		OriginalSource: nullPtrFromString(originalSource),
		StoragePrefix:  nullPtrFromString(storagePrefix),
		TotalFiles:     totalFiles,
		TotalSizeBytes: totalSizeBytes,
		Metadata:       metadata,
		CreatedAt:      createdAtTime,
		UpdatedAt:      updatedAtTime,
	}, nil
}

func nullPtrFromString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}
