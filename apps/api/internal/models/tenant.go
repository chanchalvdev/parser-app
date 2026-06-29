package models

import (
	"encoding/json"
	"time"
)

type Tenant struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	Status       string          `json:"status"`
	ContactEmail *string         `json:"contact_email,omitempty"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

