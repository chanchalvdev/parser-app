package services

import "github.com/enterprise-file-platform/api/internal/models"

type VersionService struct{}

func NewVersionService() *VersionService {
	return &VersionService{}
}

func (s *VersionService) GetVersion(env string) models.VersionResponse {
	return models.VersionResponse{
		Service:     models.APIName,
		Version:     models.APIVersion,
		Environment: env,
	}
}
