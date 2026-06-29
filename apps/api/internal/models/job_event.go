package models

import (
	"encoding/json"
	"time"
)

type JobEvent struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	JobID       string          `json:"job_id"`
	EventType   string          `json:"event_type"`
	EventMessage *string         `json:"event_message,omitempty"`
	EventDetails json.RawMessage `json:"event_details"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

