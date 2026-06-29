package models

import "time"

type ArchivePasswordRef struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	FileID         string    `json:"file_id"`
	UploadID       *string   `json:"upload_id,omitempty"`
	PasswordRefHash string   `json:"password_ref_hash"`
	Algorithm      string    `json:"algorithm"`
	IsValid        bool      `json:"is_valid"`
	Validated      bool      `json:"validated"`
	AttemptCount   int64     `json:"attempt_count"`
	LastValidated  *time.Time `json:"last_validated_at,omitempty"`
	CreatedBy      *string   `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

