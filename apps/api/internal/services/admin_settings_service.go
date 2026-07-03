package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/repositories"
)

const (
	defaultAdminMaxUploadSizeMB    = 5120
	defaultAdminMaxArchiveDepth    = 10
	defaultAdminMaxExtractedFiles  = 10000
	defaultAdminMaxExtractedSizeMB = 1024
	defaultAdminTxtSmallFileMB     = 10
	defaultAdminMaxExpansionRatio  = 20.0
	defaultAdminParserBatchSize    = 1000
	defaultAdminSearchIndexBatch   = 1000

	enabledParsersDefaultJSON = `["txt","log","csv","json","jsonl","xml","xlsx","pdf","text"]`
)

var (
	ErrInvalidSettingsRequest = errors.New("invalid settings request")
)

type AdminSettingsService struct {
	settingsRepo repositories.SettingsRepository
	auditLogRepo repositories.AuditLogRepository
}

type AdminSettingsResponse struct {
	TenantID            string   `json:"tenant_id"`
	MaxUploadSizeMB     float64  `json:"max_upload_size_mb"`
	MaxArchiveDepth     int64    `json:"max_archive_depth"`
	MaxExtractedFiles   int64    `json:"max_extracted_files"`
	MaxExtractedSizeMB  int64    `json:"max_extracted_size_mb"`
	TxtSmallFileLimitMB float64  `json:"txt_small_file_limit_mb"`
	MaxExpansionRatio   float64  `json:"max_expansion_ratio"`
	EnabledParsers      []string `json:"enabled_parsers"`
	ParserBatchSize     int64    `json:"parser_batch_size"`
	SearchIndexBatchSize int64   `json:"search_index_batch_size"`
}

type UpdateAdminSettingsRequest struct {
	MaxUploadSizeMB      *float64  `json:"max_upload_size_mb"`
	MaxArchiveDepth      *int64    `json:"max_archive_depth"`
	MaxExtractedFiles    *int64    `json:"max_extracted_files"`
	MaxExtractedSizeMB   *int64    `json:"max_extracted_size_mb"`
	TxtSmallFileLimitMB  *float64  `json:"txt_small_file_limit_mb"`
	MaxExpansionRatio    *float64  `json:"max_expansion_ratio"`
	EnabledParsers       *[]string `json:"enabled_parsers"`
	ParserBatchSize      *int64    `json:"parser_batch_size"`
	SearchIndexBatchSize *int64    `json:"search_index_batch_size"`
}

func NewAdminSettingsService(settingsRepo repositories.SettingsRepository, auditLogRepo repositories.AuditLogRepository) *AdminSettingsService {
	return &AdminSettingsService{
		settingsRepo: settingsRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *AdminSettingsService) GetSettings(ctx context.Context, tenantID string) (*AdminSettingsResponse, error) {
	if s.settingsRepo == nil {
		return nil, fmt.Errorf("admin settings service is not initialized")
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	maxUploadSizeMB, err := s.readFloatSetting(ctx, tenantID, "max_upload_size_mb", defaultAdminMaxUploadSizeMB)
	if err != nil {
		return nil, err
	}

	maxArchiveDepth, err := s.readIntSetting(ctx, tenantID, "max_archive_depth", defaultAdminMaxArchiveDepth)
	if err != nil {
		return nil, err
	}

	maxExtractedFiles, err := s.readIntSetting(ctx, tenantID, "max_extracted_files", defaultAdminMaxExtractedFiles)
	if err != nil {
		return nil, err
	}

	maxExtractedSizeMB, err := s.readIntSetting(ctx, tenantID, "max_extracted_size_mb", defaultAdminMaxExtractedSizeMB)
	if err != nil {
		return nil, err
	}

	txtSmallFileLimitMB, err := s.readFloatSetting(ctx, tenantID, "txt_small_file_limit_mb", defaultAdminTxtSmallFileMB)
	if err != nil {
		return nil, err
	}

	maxExpansionRatio, err := s.readFloatSetting(ctx, tenantID, "max_expansion_ratio", defaultAdminMaxExpansionRatio)
	if err != nil {
		return nil, err
	}

	enabledParsers, err := s.readStringSliceSetting(
		ctx,
		tenantID,
		"enabled_parsers",
		parseEnabledParsers(enabledParsersDefaultJSON),
	)
	if err != nil {
		return nil, err
	}

	parserBatchSize, err := s.readIntSetting(ctx, tenantID, "parser_batch_size", defaultAdminParserBatchSize)
	if err != nil {
		return nil, err
	}

	searchIndexBatchSize, err := s.readIntSetting(ctx, tenantID, "search_index_batch_size", defaultAdminSearchIndexBatch)
	if err != nil {
		return nil, err
	}

	return &AdminSettingsResponse{
		TenantID:             tenantID,
		MaxUploadSizeMB:      maxUploadSizeMB,
		MaxArchiveDepth:      maxArchiveDepth,
		MaxExtractedFiles:    maxExtractedFiles,
		MaxExtractedSizeMB:   maxExtractedSizeMB,
		TxtSmallFileLimitMB:  txtSmallFileLimitMB,
		MaxExpansionRatio:    maxExpansionRatio,
		EnabledParsers:       enabledParsers,
		ParserBatchSize:      parserBatchSize,
		SearchIndexBatchSize: searchIndexBatchSize,
	}, nil
}

func (s *AdminSettingsService) UpdateSettings(
	ctx context.Context,
	tenantID string,
	req UpdateAdminSettingsRequest,
	actorUserID *string,
	ipAddress *string,
	userAgent *string,
) (*AdminSettingsResponse, error) {
	if s.settingsRepo == nil || s.auditLogRepo == nil {
		return nil, fmt.Errorf("admin settings service is not fully initialized")
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	updates := make(map[string]any)

	if req.MaxUploadSizeMB != nil {
		if *req.MaxUploadSizeMB <= 0 {
			return nil, fmt.Errorf("%w: max_upload_size_mb must be greater than 0", ErrInvalidSettingsRequest)
		}
		updates["max_upload_size_mb"] = *req.MaxUploadSizeMB
	}

	if req.MaxArchiveDepth != nil {
		if *req.MaxArchiveDepth < 0 {
			return nil, fmt.Errorf("%w: max_archive_depth must be non-negative", ErrInvalidSettingsRequest)
		}
		updates["max_archive_depth"] = *req.MaxArchiveDepth
	}

	if req.MaxExtractedFiles != nil {
		if *req.MaxExtractedFiles <= 0 {
			return nil, fmt.Errorf("%w: max_extracted_files must be greater than 0", ErrInvalidSettingsRequest)
		}
		updates["max_extracted_files"] = *req.MaxExtractedFiles
	}

	if req.MaxExtractedSizeMB != nil {
		if *req.MaxExtractedSizeMB < 0 {
			return nil, fmt.Errorf("%w: max_extracted_size_mb must be non-negative", ErrInvalidSettingsRequest)
		}
		updates["max_extracted_size_mb"] = *req.MaxExtractedSizeMB
	}

	if req.TxtSmallFileLimitMB != nil {
		if *req.TxtSmallFileLimitMB < 0 {
			return nil, fmt.Errorf("%w: txt_small_file_limit_mb must be non-negative", ErrInvalidSettingsRequest)
		}
		updates["txt_small_file_limit_mb"] = *req.TxtSmallFileLimitMB
	}

	if req.MaxExpansionRatio != nil {
		if *req.MaxExpansionRatio <= 0 {
			return nil, fmt.Errorf("%w: max_expansion_ratio must be greater than 0", ErrInvalidSettingsRequest)
		}
		updates["max_expansion_ratio"] = *req.MaxExpansionRatio
	}

	if req.EnabledParsers != nil {
		parsers := normalizedStringSlice(*req.EnabledParsers)
		if len(parsers) == 0 {
			return nil, fmt.Errorf("%w: enabled_parsers must not be empty", ErrInvalidSettingsRequest)
		}
		updates["enabled_parsers"] = parsers
	}

	if req.ParserBatchSize != nil {
		if *req.ParserBatchSize <= 0 {
			return nil, fmt.Errorf("%w: parser_batch_size must be greater than 0", ErrInvalidSettingsRequest)
		}
		updates["parser_batch_size"] = *req.ParserBatchSize
	}

	if req.SearchIndexBatchSize != nil {
		if *req.SearchIndexBatchSize <= 0 {
			return nil, fmt.Errorf("%w: search_index_batch_size must be greater than 0", ErrInvalidSettingsRequest)
		}
		updates["search_index_batch_size"] = *req.SearchIndexBatchSize
	}

	if len(updates) == 0 {
		return s.GetSettings(ctx, tenantID)
	}

	for key, rawValue := range updates {
		value, err := json.Marshal(rawValue)
		if err != nil {
			return nil, fmt.Errorf("marshal setting value: %w", err)
		}

		_, err = s.settingsRepo.UpsertSetting(ctx, &models.SystemSetting{
			TenantID:     tenantID,
			SettingKey:   key,
			SettingValue: value,
			SettingType:  "json",
			Description:  ptrString("admin-config-managed"),
			IsSecret:     false,
			UpdatedBy:    actorUserID,
		})
		if err != nil {
			return nil, fmt.Errorf("update setting %s: %w", key, err)
		}
	}

	if err := s.createSettingsAuditLog(ctx, tenantID, actorUserID, updates, ipAddress, userAgent); err != nil {
		return nil, fmt.Errorf("create settings audit log: %w", err)
	}

	return s.GetSettings(ctx, tenantID)
}

func (s *AdminSettingsService) readFloatSetting(ctx context.Context, tenantID, key string, defaultValue float64) (float64, error) {
	setting, err := s.settingsRepo.GetSetting(ctx, tenantID, key)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return defaultValue, nil
		}
		return 0, fmt.Errorf("read setting %s: %w", key, err)
	}

	value, err := parseNumericJSON(setting.SettingValue)
	if err != nil {
		return 0, fmt.Errorf("parse setting %s: %w", key, err)
	}
	return value, nil
}

func (s *AdminSettingsService) readIntSetting(ctx context.Context, tenantID, key string, defaultValue int64) (int64, error) {
	value, err := s.readFloatSetting(ctx, tenantID, key, float64(defaultValue))
	if err != nil {
		return 0, err
	}

	if math.Mod(value, 1) != 0 {
		return 0, fmt.Errorf("invalid integer setting %s: %v", key, value)
	}
	return int64(value), nil
}

func (s *AdminSettingsService) readStringSliceSetting(ctx context.Context, tenantID, key string, defaultValue []string) ([]string, error) {
	setting, err := s.settingsRepo.GetSetting(ctx, tenantID, key)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return defaultValue, nil
		}
		return nil, fmt.Errorf("read setting %s: %w", key, err)
	}

	var parsed []string
	if err := json.Unmarshal(setting.SettingValue, &parsed); err != nil {
		return nil, fmt.Errorf("parse setting %s: %w", key, err)
	}
	if len(parsed) == 0 {
		return defaultValue, nil
	}

	return normalizedStringSlice(parsed), nil
}

func (s *AdminSettingsService) createSettingsAuditLog(
	ctx context.Context,
	tenantID string,
	actorUserID *string,
	updates map[string]any,
	ipAddress *string,
	userAgent *string,
) error {
	details, err := json.Marshal(map[string]any{
		"tenant_id":         tenantID,
		"updated_settings":  updates,
		"updated_count":     len(updates),
		"occurred_at":       time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	_, err = s.auditLogRepo.CreateAuditLog(ctx, &models.AuditLog{
		TenantID:   tenantID,
		ActorUserID: func() *string {
			if actorUserID != nil && strings.TrimSpace(*actorUserID) != "" {
				return actorUserID
			}
			uid := defaultUserID
			return &uid
		}(),
		Action:     "settings.changed",
		EntityType: ptrString("system_settings"),
		Details:    details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return err
}

func parseEnabledParsers(raw string) []string {
	if raw == "" {
		return []string{}
	}

	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return []string{}
	}
	return normalizedStringSlice(parsed)
}

func normalizedStringSlice(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))

	for _, value := range raw {
		parsed := strings.TrimSpace(strings.ToLower(value))
		if parsed == "" {
			continue
		}
		if _, exists := seen[parsed]; exists {
			continue
		}
		seen[parsed] = struct{}{}
		result = append(result, parsed)
	}
	return result
}
