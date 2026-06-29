package models

import (
	"time"
)

type IngestionJob struct {
	ID              string   `json:"id"`
	TenantID        string   `json:"tenant_id"`
	RootFileID      string   `json:"root_file_id"`
	Status          string   `json:"status"`
	CurrentStage    *string  `json:"current_stage,omitempty"`
	ProgressPercent float64  `json:"progress_percent"`
	RetryCount      int      `json:"retry_count"`
	ErrorCode       *string  `json:"error_code,omitempty"`
	ErrorMessage    *string  `json:"error_message,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

