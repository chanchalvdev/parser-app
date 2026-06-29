package models

import (
\t"encoding/json"
\t"time"
)

type JobStatus string

const (
\tJobStatusReceived          JobStatus = "received"
\tJobStatusQueued            JobStatus = "queued"
\tJobStatusProcessing        JobStatus = "processing"
\tJobStatusCompleted         JobStatus = "completed"
\tJobStatusFailed            JobStatus = "failed"
\tJobStatusPasswordRequired  JobStatus = "password_required"
\tJobStatusRetryRequested    JobStatus = "retry_requested"
)

type Upload struct {
\tID              string    `json:"id"`
\tFilename        string    `json:"filename"`
\tOriginalName    string    `json:"original_name"`
\tStorageKey      string    `json:"storage_key"`
\tUploaderID      string    `json:"uploader_id"`
\tStatus          string    `json:"status"`
\tContentType     string    `json:"content_type"`
\tSizeBytes       int64     `json:"size_bytes"`
\tCreatedAt       time.Time `json:"created_at"`
\tUpdatedAt       time.Time `json:"updated_at"`
\tErrorMessage    string    `json:"error_message"`
\tHasPasswordHint bool      `json:"has_password_hint"`
}

type Job struct {
\tID           string    `json:"id"`
\tUploadID     string    `json:"upload_id"`
\tStatus       string    `json:"status"`
\tStage        string    `json:"stage"`
\tAttemptCount int       `json:"attempt_count"`
\tErrorMessage string    `json:"error_message"`
\tCreatedAt    time.Time `json:"created_at"`
\tUpdatedAt    time.Time `json:"updated_at"`
\tFinishedAt   *time.Time `json:"finished_at"`
}

type FileRecord struct {
\tID          string          `json:"id"`
\tUploadID    string          `json:"upload_id"`
\tJobID       string          `json:"job_id"`
\tParentID    *string         `json:"parent_id"`
\tPath        string          `json:"path"`
\tFilename    string          `json:"filename"`
\tKind        string          `json:"kind"`
\tMIMEType    string          `json:"mime_type"`
\tSizeBytes   int64           `json:"size_bytes"`
\tSHA256      string          `json:"sha256"`
\tMetadata    json.RawMessage `json:"metadata"`
\tCreatedAt   time.Time       `json:"created_at"`
\tChildren    []*FileRecord   `json:"children,omitempty"`
}

type ParsedRecord struct {
\tID           string          `json:"id"`
\tFileID       string          `json:"file_id"`
\tUploadID     string          `json:"upload_id"`
\tJobID        string          `json:"job_id"`
\tFilePath     string          `json:"file_path"`
\tSourceFormat string          `json:"source_format"`
\tContent      string          `json:"content"`
\tMetadata     json.RawMessage `json:"metadata"`
\tCreatedAt    time.Time       `json:"created_at"`
}

type JobPayload struct {
\tJobID      string `json:"job_id"`
\tUploadID   string `json:"upload_id"`
\tObjectKey  string `json:"object_key"`
\tFilename   string `json:"filename"`
\tPassword   string `json:"password"`
\tAttempts   int    `json:"attempts"`
}

type SearchResult struct {
\tFileID    string `json:"file_id"`
\tUploadID  string `json:"upload_id"`
\tJobID     string `json:"job_id"`
\tFilename  string `json:"filename"`
\tPath      string `json:"path"`
\tContent   string `json:"content"`
\tScore     float64 `json:"score"`
\tIndexID   string `json:"index_id"`
}

type DashboardSummary struct {
\tJobsByStatus map[string]int `json:"jobs_by_status"`
\tUploads      int            `json:"uploads"`
\tFiles        int            `json:"files"`
\tParsedCount  int            `json:"parsed_records"`
\tRecentJobs   []Job          `json:"recent_jobs"`
}

type AuditLog struct {
\tID        int             `json:"id"`
\tUploadID  string          `json:"upload_id"`
\tJobID     string          `json:"job_id"`
\tActor     string          `json:"actor"`
\tEvent     string          `json:"event"`
\tDetails   json.RawMessage `json:"details"`
\tCreatedAt time.Time       `json:"created_at"`
}
