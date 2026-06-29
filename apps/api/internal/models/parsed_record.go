package models

import (
	"encoding/json"
	"time"
)

type ParsedRecord struct {
	ID                string          `json:"id"`
	TenantID          string          `json:"tenant_id"`
	FileID            string          `json:"file_id"`
	JobID             string          `json:"job_id"`
	RecordType        *string         `json:"record_type,omitempty"`
	RecordNumber      *int64          `json:"record_number,omitempty"`
	LineNumber        *int64          `json:"line_number,omitempty"`
	ChunkNumber       *int64          `json:"chunk_number,omitempty"`
	StartLine         *int64          `json:"start_line,omitempty"`
	EndLine           *int64          `json:"end_line,omitempty"`
	ContentText       *string         `json:"content_text,omitempty"`
	StructuredData    json.RawMessage `json:"structured_data"`
	ExtractedEntities json.RawMessage `json:"extracted_entities"`
	EventTimestamp    *time.Time      `json:"event_timestamp,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

