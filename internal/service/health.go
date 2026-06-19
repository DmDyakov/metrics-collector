package service

import (
	"context"
)

//go:generate mockgen -destination=mocks/mock_health_repository.go -package=mocks . HealthRepository
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	repo HealthRepository
}

func NewHealthService(repo HealthRepository) *HealthService {
	return &HealthService{repo: repo}
}

func (svc *HealthService) Ping(ctx context.Context) error {
	return svc.repo.Ping(ctx)
}
