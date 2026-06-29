package services

import (
	"context"

	"github.com/enterprise-file-platform/api/internal/repositories"
)

type HealthService struct {
	repo repositories.HealthRepository
}

func NewHealthService(repo repositories.HealthRepository) *HealthService {
	return &HealthService{repo: repo}
}

func (s *HealthService) Ready(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
