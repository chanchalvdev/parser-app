package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsRepository interface {
	GetSetting(ctx context.Context, tenantID string, settingKey string) (*models.SystemSetting, error)
	UpsertSetting(ctx context.Context, setting *models.SystemSetting) (*models.SystemSetting, error)
}

type PostgresSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewSettingsRepository(pool *pgxpool.Pool) *PostgresSettingsRepository {
	return &PostgresSettingsRepository{pool: pool}
}

func (r *PostgresSettingsRepository) GetSetting(ctx context.Context, tenantID string, settingKey string) (*models.SystemSetting, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" || settingKey == "" {
		return nil, fmt.Errorf("tenant_id and setting_key are required")
	}

	query := `
		SELECT id, tenant_id, setting_key, setting_value, setting_type, description, is_secret, updated_by, created_at, updated_at
		FROM system_settings
		WHERE tenant_id = $1 AND setting_key = $2
	`
	row := r.pool.QueryRow(ctx, query, tenantID, settingKey)
	setting, err := scanSystemSetting(row)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return setting, nil
}

func (r *PostgresSettingsRepository) UpsertSetting(ctx context.Context, setting *models.SystemSetting) (*models.SystemSetting, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, fmt.Errorf("setting is required")
	}
	if setting.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if setting.SettingKey == "" {
		return nil, fmt.Errorf("setting_key is required")
	}
	if setting.ID == "" {
		setting.ID = uuid.NewString()
	}
	if setting.SettingType == "" {
		setting.SettingType = "json"
	}

	value, err := asJSON(setting.SettingValue)
	if err != nil {
		return nil, fmt.Errorf("setting value: %w", err)
	}

	query := `
		INSERT INTO system_settings (
			id, tenant_id, setting_key, setting_value, setting_type, description, is_secret, updated_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, setting_key)
		DO UPDATE
		SET setting_value = EXCLUDED.setting_value,
			setting_type = EXCLUDED.setting_type,
			description = EXCLUDED.description,
			is_secret = EXCLUDED.is_secret,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
		RETURNING id, tenant_id, setting_key, setting_value, setting_type, description, is_secret, updated_by, created_at, updated_at
	`
	row := r.pool.QueryRow(
		ctx,
		query,
		setting.ID,
		setting.TenantID,
		setting.SettingKey,
		value,
		setting.SettingType,
		nullString(setting.Description),
		setting.IsSecret,
		nullString(setting.UpdatedBy),
	)
	return scanSystemSetting(row)
}

func scanSystemSetting(scanner rowScanner) (*models.SystemSetting, error) {
	var (
		id, tenantID, settingKey, settingType string
		settingValue                         []byte
		description, updatedBy                sql.NullString
		isSecret                             bool
		createdAt, updatedAt                 sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&tenantID,
		&settingKey,
		&settingValue,
		&settingType,
		&description,
		&isSecret,
		&updatedBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, mapNoRows(err)
	}

	value, err := scanJSON(settingValue)
	if err != nil {
		return nil, err
	}
	ca := createdAt.Time
	ua := updatedAt.Time
	if !createdAt.Valid {
		ca = time.Time{}
	}
	if !updatedAt.Valid {
		ua = time.Time{}
	}

	return &models.SystemSetting{
		ID:           id,
		TenantID:     tenantID,
		SettingKey:   settingKey,
		SettingValue: value,
		SettingType:  settingType,
		Description:  nullPtrFromString(description),
		IsSecret:     isSecret,
		UpdatedBy:    nullPtrFromString(updatedBy),
		CreatedAt:    ca,
		UpdatedAt:    ua,
	}, nil
}
