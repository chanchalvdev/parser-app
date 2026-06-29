package models

import (
	"time"
)

type File struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	ParentFileID       *string         `json:"parent_file_id,omitempty"`
	UploadID           *string         `json:"upload_id,omitempty"`
	OriginalName       string          `json:"original_name"`
	NormalizedName     *string         `json:"normalized_name,omitempty"`
	Extension          *string         `json:"extension,omitempty"`
	DetectedMimeType   *string         `json:"detected_mime_type,omitempty"`
	DetectedFileType   *string         `json:"detected_file_type,omitempty"`
	StoragePath        string          `json:"storage_path"`
	SizeBytes          int64           `json:"size_bytes"`
	Sha256Hash         *string         `json:"sha256_hash,omitempty"`
	Depth              int             `json:"depth"`
	IsArchive          bool            `json:"is_archive"`
	IsPasswordProtected bool           `json:"is_password_protected"`
	ProcessingStatus   string          `json:"processing_status"`
	CreatedBy          *string         `json:"created_by,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
