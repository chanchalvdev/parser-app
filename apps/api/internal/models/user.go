package models

import (
	"time"
)

type User struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	Email        string `json:"email"`
	PasswordHash *string `json:"password_hash,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
