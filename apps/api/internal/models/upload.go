package models

import (
	"encoding/json"
	"time"
)

type Upload struct {
	ID              string          `json:"id"`
	TenantID        string          `json:"tenant_id"`
	InitiatedBy     *string         `json:"initiated_by,omitempty"`
	Status          string          `json:"status"`
	OriginalSource  *string         `json:"original_source,omitempty"`
	StoragePrefix   *string         `json:"storage_prefix,omitempty"`
	RequestedBy     *string         `json:"requested_by,omitempty"`
	TotalFiles      int             `json:"total_files"`
	TotalSizeBytes  int64           `json:"total_size_bytes"`
	Metadata        json.RawMessage `json:"metadata"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

