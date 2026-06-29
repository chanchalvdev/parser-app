package models

import (
	"encoding/json"
	"time"
)

type SystemSetting struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	SettingKey   string          `json:"setting_key"`
	SettingValue json.RawMessage `json:"setting_value"`
	SettingType  string          `json:"setting_type"`
	Description  *string         `json:"description,omitempty"`
	IsSecret     bool            `json:"is_secret"`
	UpdatedBy    *string         `json:"updated_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
