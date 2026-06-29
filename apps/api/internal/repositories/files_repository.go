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

type FileRepository interface {
	CreateFile(ctx context.Context, file *models.File) (*models.File, error)
	GetFileByID(ctx context.Context, id string) (*models.File, error)
	ListFiles(ctx context.Context, tenantID string) ([]*models.File, error)
	ListFilesFiltered(
		ctx context.Context,
		tenantID string,
		processingStatus string,
		extension string,
		detectedFileType string,
		limit int,
		offset int,
	) ([]*models.File, error)
	CountFiles(
		ctx context.Context,
		tenantID string,
		processingStatus string,
		extension string,
		detectedFileType string,
	) (int64, error)
	ListChildren(ctx context.Context, parentFileID string) ([]*models.File, error)
	ListChildrenPaginated(ctx context.Context, parentFileID string, limit int, offset int) ([]*models.File, error)
	CountChildren(ctx context.Context, parentFileID string) (int64, error)
	UpdateFileStatus(ctx context.Context, id string, status string) error
}

type PostgresFileRepository struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *PostgresFileRepository {
	return &PostgresFileRepository{pool: pool}
}

func (r *PostgresFileRepository) CreateFile(ctx context.Context, file *models.File) (*models.File, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}
	if file.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if file.ID == "" {
		file.ID = uuid.NewString()
	}
	if file.OriginalName == "" {
		return nil, fmt.Errorf("original_name is required")
	}
	if file.StoragePath == "" {
		return nil, fmt.Errorf("storage_path is required")
	}
	if file.ProcessingStatus == "" {
		file.ProcessingStatus = "pending"
	}

	query := `
		INSERT INTO files (
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by, created_at, updated_at
	`
	row := r.pool.QueryRow(
		ctx,
		query,
		file.ID,
		file.TenantID,
		nullString(file.ParentFileID),
		nullString(file.UploadID),
		file.OriginalName,
		nullString(file.NormalizedName),
		nullString(file.Extension),
		nullString(file.DetectedMimeType),
		nullString(file.DetectedFileType),
		file.StoragePath,
		file.SizeBytes,
		nullString(file.Sha256Hash),
		file.Depth,
		file.IsArchive,
		file.IsPasswordProtected,
		file.ProcessingStatus,
		nullString(file.CreatedBy),
	)

	return scanFile(row)
}

func (r *PostgresFileRepository) GetFileByID(ctx context.Context, id string) (*models.File, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("file id is required")
	}
	query := `
		SELECT
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by, created_at, updated_at
		FROM files
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	file, err := scanFile(row)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return file, nil
}

func (r *PostgresFileRepository) ListFiles(ctx context.Context, tenantID string) ([]*models.File, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	query := `
		SELECT
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by, created_at, updated_at
		FROM files
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.File, 0)
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresFileRepository) ListFilesFiltered(
	ctx context.Context,
	tenantID string,
	processingStatus string,
	extension string,
	detectedFileType string,
	limit int,
	offset int,
) ([]*models.File, error) {
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

	var (
		query      string
		count      int
		queryArgs  = []any{tenantID}
		conditions = "WHERE tenant_id = $1"
	)

	if strings.TrimSpace(processingStatus) != "" {
		queryArgs = append(queryArgs, strings.TrimSpace(processingStatus))
		count++
		conditions += fmt.Sprintf(" AND processing_status = $%d", count+1)
	}
	if strings.TrimSpace(extension) != "" {
		queryArgs = append(queryArgs, strings.TrimSpace(extension))
		count++
		conditions += fmt.Sprintf(" AND extension = $%d", count+1)
	}
	if strings.TrimSpace(detectedFileType) != "" {
		queryArgs = append(queryArgs, strings.TrimSpace(detectedFileType))
		count++
		conditions += fmt.Sprintf(" AND detected_file_type = $%d", count+1)
	}

	query = fmt.Sprintf(`
		SELECT
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by, created_at, updated_at
		FROM files
		%s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, conditions, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.File, 0)
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresFileRepository) CountFiles(
	ctx context.Context,
	tenantID string,
	processingStatus string,
	extension string,
	detectedFileType string,
) (int64, error) {
	if err := validatePool(r.pool); err != nil {
		return 0, err
	}
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}

	var (
		query      string
		count      int
		args       = []any{tenantID}
		conditions = "WHERE tenant_id = $1"
	)
	if strings.TrimSpace(processingStatus) != "" {
		args = append(args, strings.TrimSpace(processingStatus))
		count++
		conditions += fmt.Sprintf(" AND processing_status = $%d", count+1)
	}
	if strings.TrimSpace(extension) != "" {
		args = append(args, strings.TrimSpace(extension))
		count++
		conditions += fmt.Sprintf(" AND extension = $%d", count+1)
	}
	if strings.TrimSpace(detectedFileType) != "" {
		args = append(args, strings.TrimSpace(detectedFileType))
		count++
		conditions += fmt.Sprintf(" AND detected_file_type = $%d", count+1)
	}

	query = fmt.Sprintf(`
		SELECT COUNT(*)
		FROM files
		%s
	`, conditions)

	var total int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PostgresFileRepository) ListChildren(ctx context.Context, parentFileID string) ([]*models.File, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if parentFileID == "" {
		return nil, fmt.Errorf("parent_file_id is required")
	}
	query := `
		SELECT
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by, created_at, updated_at
		FROM files
		WHERE parent_file_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, parentFileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.File, 0)
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresFileRepository) ListChildrenPaginated(ctx context.Context, parentFileID string, limit int, offset int) ([]*models.File, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if parentFileID == "" {
		return nil, fmt.Errorf("parent_file_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := `
		SELECT
			id, tenant_id, parent_file_id, upload_id, original_name, normalized_name, extension,
			detected_mime_type, detected_file_type, storage_path, size_bytes, sha256_hash, depth,
			is_archive, is_password_protected, processing_status, created_by, created_at, updated_at
		FROM files
		WHERE parent_file_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, parentFileID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.File, 0)
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresFileRepository) CountChildren(ctx context.Context, parentFileID string) (int64, error) {
	if err := validatePool(r.pool); err != nil {
		return 0, err
	}
	if parentFileID == "" {
		return 0, fmt.Errorf("parent_file_id is required")
	}
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM files WHERE parent_file_id = $1`, parentFileID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PostgresFileRepository) UpdateFileStatus(ctx context.Context, id string, status string) error {
	if err := validatePool(r.pool); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("file id is required")
	}
	if status == "" {
		return fmt.Errorf("status is required")
	}
	query := `
		UPDATE files
		SET processing_status = $2,
			updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return err
	}
	return rowsAffectedError(tag, "file")
}

func scanFile(scanner rowScanner) (*models.File, error) {
	var (
		id, tenantID                                                string
		parentFileID, uploadID, normalizedName, extension           sql.NullString
		detectedMimeType, detectedFileType                          sql.NullString
		storagePath                                                 string
		sizeBytes                                                   int64
		sha256Hash                                                  sql.NullString
		depth                                                       int
		isArchive, isPasswordProtected                              bool
		processingStatus                                            string
		createdBy                                                  sql.NullString
		originalName                                                string
		createdAt, updatedAt                                        sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&tenantID,
		&parentFileID,
		&uploadID,
		&originalName,
		&normalizedName,
		&extension,
		&detectedMimeType,
		&detectedFileType,
		&storagePath,
		&sizeBytes,
		&sha256Hash,
		&depth,
		&isArchive,
		&isPasswordProtected,
		&processingStatus,
		&createdBy,
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

	return &models.File{
		ID:                 id,
		TenantID:           tenantID,
		ParentFileID:       nullPtrFromString(parentFileID),
		UploadID:           nullPtrFromString(uploadID),
		OriginalName:       originalName,
		NormalizedName:     nullPtrFromString(normalizedName),
		Extension:          nullPtrFromString(extension),
		DetectedMimeType:   nullPtrFromString(detectedMimeType),
		DetectedFileType:   nullPtrFromString(detectedFileType),
		StoragePath:        storagePath,
		SizeBytes:          sizeBytes,
		Sha256Hash:         nullPtrFromString(sha256Hash),
		Depth:              depth,
		IsArchive:          isArchive,
		IsPasswordProtected: isPasswordProtected,
		ProcessingStatus:   processingStatus,
		CreatedBy:          nullPtrFromString(createdBy),
		CreatedAt:          ca,
		UpdatedAt:          ua,
	}, nil
}
