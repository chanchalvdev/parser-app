package services

import (
	"context"
	"testing"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
)

type fakeDashboardRepository struct {
	summary             *models.DashboardSummary
	volume              []*models.UploadVolumeBucket
	distribution        []*models.DashboardBucket
	errorBreakdown      *models.DashboardErrorBreakdown
	entities            map[string][]*models.DashboardBucket
	processingDuration   *models.DashboardProcessingDuration
}

func (f *fakeDashboardRepository) GetDashboardSummary(_ context.Context, tenantID string) (*models.DashboardSummary, error) {
	if f.summary == nil {
		f.summary = &models.DashboardSummary{TenantID: tenantID}
	}
	f.summary.TenantID = tenantID
	return f.summary, nil
}

func (f *fakeDashboardRepository) GetDashboardFileTypeDistribution(_ context.Context, _ string, _ int) ([]*models.DashboardBucket, error) {
	return f.distribution, nil
}

func (f *fakeDashboardRepository) GetDashboardProcessingStatusDistribution(_ context.Context, _ string, _ int) ([]*models.DashboardBucket, error) {
	return f.distribution, nil
}

func (f *fakeDashboardRepository) GetDashboardUploadVolume(_ context.Context, _ string, _, _ time.Time) ([]*models.UploadVolumeBucket, error) {
	return f.volume, nil
}

func (f *fakeDashboardRepository) GetDashboardErrorBreakdown(_ context.Context, _ string, _ int) (*models.DashboardErrorBreakdown, error) {
	return f.errorBreakdown, nil
}

func (f *fakeDashboardRepository) GetDashboardTopEntities(_ context.Context, _ string, _ int) (map[string][]*models.DashboardBucket, error) {
	return f.entities, nil
}

func (f *fakeDashboardRepository) GetDashboardProcessingDuration(_ context.Context, tenantID string) (*models.DashboardProcessingDuration, error) {
	if f.processingDuration == nil {
		f.processingDuration = &models.DashboardProcessingDuration{TenantID: tenantID}
	}
	f.processingDuration.TenantID = tenantID
	return f.processingDuration, nil
}

func TestDashboardServiceGetSummaryUsesDefaultTenant(t *testing.T) {
	repo := &fakeDashboardRepository{
		summary: &models.DashboardSummary{TenantID: "tenant"},
	}
	service := NewDashboardService(repo)

	result, err := service.GetSummary(context.Background(), DashboardSummaryRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TenantID != defaultTenantID {
		t.Fatalf("expected tenant %s, got %s", defaultTenantID, result.TenantID)
	}
}

func TestDashboardServiceGetUploadVolumeClampsDays(t *testing.T) {
	repo := &fakeDashboardRepository{}
	service := NewDashboardService(repo)

	result, err := service.GetUploadVolume(context.Background(), DashboardUploadVolumeRequest{
		TenantID: "tenant-1",
		Days:     maxUploadVolumeDays + 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TenantID != "tenant-1" {
		t.Fatalf("expected tenant-1, got %s", result.TenantID)
	}
	if result.Days != maxUploadVolumeDays {
		t.Fatalf("expected days %d, got %d", maxUploadVolumeDays, result.Days)
	}
}
