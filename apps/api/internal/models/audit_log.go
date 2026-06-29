package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenant_id"`
	ActorUserID *string        `json:"actor_user_id,omitempty"`
	Action     string          `json:"action"`
	EntityType *string         `json:"entity_type,omitempty"`
	EntityID   *string         `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	UserAgent  *string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

