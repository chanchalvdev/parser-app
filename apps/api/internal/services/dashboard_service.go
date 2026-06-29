package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/repositories"
)

const (
	defaultDashboardListLimit = 25
	maxDashboardListLimit     = 200
	defaultUploadVolumeDays   = 7
	maxUploadVolumeDays       = 365
)

type DashboardService struct {
	dashboardRepo repositories.DashboardRepository
}

type DashboardSummaryRequest struct {
	TenantID string
}

type DashboardDistributionRequest struct {
	TenantID string
	Limit    int
}

type DashboardUploadVolumeRequest struct {
	TenantID string
	Days     int
}

func NewDashboardService(repo repositories.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: repo}
}

func (s *DashboardService) GetSummary(ctx context.Context, req DashboardSummaryRequest) (*models.DashboardSummary, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	return s.dashboardRepo.GetDashboardSummary(ctx, tenantID)
}

func (s *DashboardService) GetFileTypeDistribution(ctx context.Context, req DashboardDistributionRequest) (*models.DashboardDistribution, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	limit := normalizeDashboardLimit(req.Limit)
	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	buckets, err := s.dashboardRepo.GetDashboardFileTypeDistribution(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}

	return &models.DashboardDistribution{
		TenantID: tenantID,
		Buckets:  buckets,
	}, nil
}

func (s *DashboardService) GetProcessingStatusDistribution(ctx context.Context, req DashboardDistributionRequest) (*models.DashboardDistribution, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	limit := normalizeDashboardLimit(req.Limit)
	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	buckets, err := s.dashboardRepo.GetDashboardProcessingStatusDistribution(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}

	return &models.DashboardDistribution{
		TenantID: tenantID,
		Buckets:  buckets,
	}, nil
}

func (s *DashboardService) GetUploadVolume(ctx context.Context, req DashboardUploadVolumeRequest) (*models.DashboardUploadVolume, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	days := req.Days
	if days <= 0 {
		days = defaultUploadVolumeDays
	}
	if days > maxUploadVolumeDays {
		days = maxUploadVolumeDays
	}

	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -(days - 1))

	series, err := s.dashboardRepo.GetDashboardUploadVolume(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &models.DashboardUploadVolume{
		TenantID: tenantID,
		Unit:     "day",
		Days:     days,
		From:     startDate.Format("2006-01-02"),
		To:       endDate.Format("2006-01-02"),
		Buckets:  series,
	}, nil
}

func (s *DashboardService) GetErrorBreakdown(ctx context.Context, req DashboardDistributionRequest) (*models.DashboardErrorBreakdown, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	limit := normalizeDashboardLimit(req.Limit)
	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	breakdown, err := s.dashboardRepo.GetDashboardErrorBreakdown(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}
	return breakdown, nil
}

func (s *DashboardService) GetTopEntities(ctx context.Context, req DashboardDistributionRequest) (*models.DashboardEntities, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	limit := normalizeDashboardLimit(req.Limit)
	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	entities, err := s.dashboardRepo.GetDashboardTopEntities(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}

	return &models.DashboardEntities{
		TenantID: tenantID,
		Limit:    limit,
		Entities: entities,
	}, nil
}

func (s *DashboardService) GetProcessingDuration(ctx context.Context, req DashboardSummaryRequest) (*models.DashboardProcessingDuration, error) {
	if s.dashboardRepo == nil {
		return nil, fmt.Errorf("dashboard service is not initialized")
	}

	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	return s.dashboardRepo.GetDashboardProcessingDuration(ctx, tenantID)
}

func normalizeDashboardLimit(limit int) int {
	return normalizePageSize(limit, defaultDashboardListLimit, maxDashboardListLimit)
}
