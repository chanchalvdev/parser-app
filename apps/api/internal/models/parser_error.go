package models

import (
	"encoding/json"
	"time"
)

type ParserError struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenant_id"`
	JobID      string          `json:"job_id"`
	FileID     *string         `json:"file_id,omitempty"`
	UploadID   *string         `json:"upload_id,omitempty"`
	ErrorCode  *string         `json:"error_code,omitempty"`
	ErrorMessage string        `json:"error_message"`
	ErrorContext json.RawMessage `json:"error_context"`
	IsRetryable bool           `json:"is_retryable"`
	StackTrace *string         `json:"stack_trace,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
}

