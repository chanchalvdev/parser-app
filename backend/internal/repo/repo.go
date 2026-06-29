package repo

import (
\t"context"
\t"encoding/json"

\t"github.com/example/file-platform/backend/internal/models"
\t"github.com/jackc/pgx/v5"
\t"github.com/jackc/pgx/v5/pgxpool"
\t"github.com/google/uuid"
)

type Repository struct {
\tDB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
\treturn &Repository{DB: db}
}

func (r *Repository) CreateUpload(ctx context.Context, upload *models.Upload) (string, error) {
\tid := uuid.NewString()
\t_, err := r.DB.Exec(ctx, `
INSERT INTO uploads (
  id, filename, original_name, storage_key, uploader_id, status, content_type, size_bytes, has_password_hint, error_message
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
\t\tid, upload.Filename, upload.OriginalName, upload.StorageKey, upload.UploaderID, upload.Status, upload.ContentType, upload.SizeBytes, upload.HasPasswordHint, upload.ErrorMessage)
\treturn id, err
}

func (r *Repository) SetUploadStatus(ctx context.Context, uploadID, status, errorMsg string) error {
\t_, err := r.DB.Exec(ctx, `
UPDATE uploads SET status=$1, error_message=$2, updated_at=NOW() WHERE id=$3`,
\t\tstatus, errorMsg, uploadID)
\treturn err
}

func (r *Repository) GetUpload(ctx context.Context, id string) (*models.Upload, error) {
\trow := r.DB.QueryRow(ctx, `
SELECT id, filename, original_name, storage_key, uploader_id, status, content_type, size_bytes, created_at, updated_at, error_message, has_password_hint
FROM uploads WHERE id=$1`, id)
\tvar u models.Upload
\tif err := row.Scan(&u.ID, &u.Filename, &u.OriginalName, &u.StorageKey, &u.UploaderID, &u.Status,
\t\t&u.ContentType, &u.SizeBytes, &u.CreatedAt, &u.UpdatedAt, &u.ErrorMessage, &u.HasPasswordHint); err != nil {
\t\treturn nil, err
\t}
\treturn &u, nil
}

func (r *Repository) ListUploads(ctx context.Context, limit, offset int) ([]models.Upload, error) {
\trows, err := r.DB.Query(ctx, `
SELECT id, filename, original_name, storage_key, uploader_id, status, content_type, size_bytes, created_at, updated_at, error_message, has_password_hint
FROM uploads ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
\tif err != nil {
\t\treturn nil, err
\t}
\tdefer rows.Close()
\titems := make([]models.Upload, 0)
\tfor rows.Next() {
\t\tvar u models.Upload
\t\tif err := rows.Scan(&u.ID, &u.Filename, &u.OriginalName, &u.StorageKey, &u.UploaderID, &u.Status, &u.ContentType, &u.SizeBytes, &u.CreatedAt, &u.UpdatedAt, &u.ErrorMessage, &u.HasPasswordHint); err != nil {
\t\t\treturn nil, err
\t\t}
\t\titems = append(items, u)
\t}
\treturn items, rows.Err()
}

func (r *Repository) CreateJob(ctx context.Context, job *models.Job) (string, error) {
\tid := uuid.NewString()
\t_, err := r.DB.Exec(ctx, `
INSERT INTO jobs (id, upload_id, status, stage, attempt_count, error_message)
VALUES ($1,$2,$3,$4,$5,$6)`,
\t\tid, job.UploadID, job.Status, job.Stage, job.AttemptCount, job.ErrorMessage)
\treturn id, err
}

func (r *Repository) SetJobStatus(ctx context.Context, jobID, status, stage, errorMsg string) error {
\t_, err := r.DB.Exec(ctx, `
UPDATE jobs
SET status=$1, stage=$2, error_message=$3, updated_at=NOW(),
    attempt_count=attempt_count + CASE WHEN $1 IN ('failed','password_required') THEN 1 ELSE 0 END,
    finished_at=CASE WHEN $1 IN ('completed','failed','password_required') THEN NOW() ELSE finished_at END
WHERE id=$4`, status, stage, errorMsg, jobID)
\treturn err
}

func (r *Repository) GetJob(ctx context.Context, id string) (*models.Job, error) {
\trow := r.DB.QueryRow(ctx, `
SELECT id, upload_id, status, stage, attempt_count, error_message, created_at, updated_at, finished_at
FROM jobs WHERE id=$1`, id)
\tvar j models.Job
\tif err := row.Scan(&j.ID, &j.UploadID, &j.Status, &j.Stage, &j.AttemptCount, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt, &j.FinishedAt); err != nil {
\t\treturn nil, err
\t}
\treturn &j, nil
}

func (r *Repository) ListJobs(ctx context.Context, limit, offset int) ([]models.Job, error) {
\trows, err := r.DB.Query(ctx, `
SELECT id, upload_id, status, stage, attempt_count, error_message, created_at, updated_at, finished_at
FROM jobs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
\tif err != nil {
\t\treturn nil, err
\t}
\tdefer rows.Close()
\titems := make([]models.Job, 0)
\tfor rows.Next() {
\t\tvar j models.Job
\t\tif err := rows.Scan(&j.ID, &j.UploadID, &j.Status, &j.Stage, &j.AttemptCount, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt, &j.FinishedAt); err != nil {
\t\t\treturn nil, err
\t\t}
\t\titems = append(items, j)
\t}
\treturn items, rows.Err()
}

func (r *Repository) GetAuditLogs(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
\trows, err := r.DB.Query(ctx, `
SELECT id, upload_id, job_id, actor, event, details, created_at
FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
\tif err != nil {
\t\treturn nil, err
\t}
\tdefer rows.Close()
\titems := make([]models.AuditLog, 0)
\tfor rows.Next() {
\t\tvar item models.AuditLog
\t\tif err := rows.Scan(&item.ID, &item.UploadID, &item.JobID, &item.Actor, &item.Event, &item.Details, &item.CreatedAt); err != nil {
\t\t\treturn nil, err
\t\t}
\t\titems = append(items, item)
\t}
\treturn items, rows.Err()
}

func (r *Repository) RecordAuditEvent(ctx context.Context, uploadID, jobID, actor, event string, details any) error {
\tb, err := json.Marshal(details)
\tif err != nil {
\t\tb = []byte(`{}`)
\t}
\t_, err = r.DB.Exec(ctx, `
INSERT INTO audit_logs (upload_id, job_id, actor, event, details, created_at)
VALUES ($1,$2,$3,$4,$5,NOW())`, uploadID, jobID, actor, event, b)
\treturn err
}

func (r *Repository) AddFileRecord(ctx context.Context, file *models.FileRecord) (string, error) {
\tid := uuid.NewString()
\t_, err := r.DB.Exec(ctx, `
INSERT INTO files (
  id, upload_id, job_id, parent_id, path, filename, kind, mime_type, size_bytes, sha256, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
\t\tid, file.UploadID, file.JobID, nullable(file.ParentID), file.Path, file.Filename, file.Kind, file.MIMEType, file.SizeBytes, file.SHA256, file.Metadata)
\treturn id, err
}

func (r *Repository) InsertFileRecord(ctx context.Context, file *models.FileRecord) (string, error) {
\treturn r.AddFileRecord(ctx, file)
}

func (r *Repository) InsertParsedRecord(ctx context.Context, parsed *models.ParsedRecord) (string, error) {
\tid := uuid.NewString()
\t_, err := r.DB.Exec(ctx, `
INSERT INTO parsed_records (id, file_id, upload_id, job_id, file_path, source_format, content, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, parsed.FileID, parsed.UploadID, parsed.JobID, parsed.FilePath, parsed.SourceFormat, parsed.Content, parsed.Metadata)
\treturn id, err
}

func (r *Repository) GetFilesByUpload(ctx context.Context, uploadID string) ([]models.FileRecord, error) {
\trows, err := r.DB.Query(ctx, `
SELECT id, upload_id, job_id, COALESCE(parent_id::text, ''), path, filename, kind, mime_type, size_bytes, sha256, metadata, created_at
FROM files WHERE upload_id=$1 ORDER BY path`, uploadID)
\tif err != nil {
\t\treturn nil, err
\t}
\tdefer rows.Close()
\tfiles := make([]models.FileRecord, 0)
\tfor rows.Next() {
\t\tvar row models.FileRecord
\t\tvar parentIDStr string
\t\tif err := rows.Scan(&row.ID, &row.UploadID, &row.JobID, &parentIDStr, &row.Path, &row.Filename, &row.Kind, &row.MIMEType, &row.SizeBytes, &row.SHA256, &row.Metadata, &row.CreatedAt); err != nil {
\t\t\treturn nil, err
\t\t}
\t\tif parentIDStr != "" {
\t\t\trow.ParentID = &parentIDStr
\t\t}
\t\tfiles = append(files, row)
\t}
\tif err := rows.Err(); err != nil {
\t\treturn nil, err
\t}
\treturn files, nil
}

func (r *Repository) GetDashboardSummary(ctx context.Context) (models.DashboardSummary, error) {
\tsummary := models.DashboardSummary{
\t\tJobsByStatus: map[string]int{},
\t}
\tstatusRows, err := r.DB.Query(ctx, `SELECT status, count(*) FROM jobs GROUP BY status`)
\tif err != nil {
\t\treturn summary, err
\t}
\tfor statusRows.Next() {
\t\tvar status string
\t\tvar count int
\t\tif err := statusRows.Scan(&status, &count); err != nil {
\t\t\tstatusRows.Close()
\t\t\treturn summary, err
\t\t}
\t\tsummary.JobsByStatus[status] = count
\t}
\tstatusRows.Close()
\tif err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM uploads`).Scan(&summary.Uploads); err != nil {
\t\treturn summary, err
\t}
\tif err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM files`).Scan(&summary.Files); err != nil {
\t\treturn summary, err
\t}
\tif err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM parsed_records`).Scan(&summary.ParsedCount); err != nil {
\t\treturn summary, err
\t}
\trecentRows, err := r.DB.Query(ctx, `SELECT id, upload_id, status, stage, attempt_count, error_message, created_at, updated_at, finished_at FROM jobs ORDER BY created_at DESC LIMIT 5`)
\tif err != nil {
\t\treturn summary, err
\t}
\tfor recentRows.Next() {
\t\tvar j models.Job
\t\tif err := recentRows.Scan(&j.ID, &j.UploadID, &j.Status, &j.Stage, &j.AttemptCount, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt, &j.FinishedAt); err != nil {
\t\t\trecentRows.Close()
\t\t\treturn summary, err
\t\t}
\t\tsummary.RecentJobs = append(summary.RecentJobs, j)
\t}
\trecentRows.Close()
\treturn summary, recentRows.Err()
}

func (r *Repository) LastJobForUpload(ctx context.Context, uploadID string) (*models.Job, error) {
\trow := r.DB.QueryRow(ctx, `
SELECT id, upload_id, status, stage, attempt_count, error_message, created_at, updated_at, finished_at
FROM jobs WHERE upload_id=$1 ORDER BY created_at DESC LIMIT 1`, uploadID)
\tvar j models.Job
\tif err := row.Scan(&j.ID, &j.UploadID, &j.Status, &j.Stage, &j.AttemptCount, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt, &j.FinishedAt); err != nil {
\t\tif err == pgx.ErrNoRows {
\t\t\treturn nil, nil
\t\t}
\t\treturn nil, err
\t}
\treturn &j, nil
}

func (r *Repository) SetUploadAndJobStatus(ctx context.Context, uploadID, jobID, status, errorMessage string) error {
\t_, err := r.DB.Exec(ctx, `
UPDATE uploads SET status=$1, error_message=$2, updated_at=NOW() WHERE id=$3`, status, errorMessage, uploadID)
\tif err != nil {
\t\treturn err
\t}
\t_, err = r.DB.Exec(ctx, `
UPDATE jobs SET status=$1, error_message=$2, updated_at=NOW(), attempt_count=attempt_count+1, stage='queued' WHERE id=$3`, status, errorMessage, jobID)
\treturn err
}

func nullable(v *string) any {
\tif v == nil || *v == "" {
\t\treturn nil
\t}
\treturn *v
}
