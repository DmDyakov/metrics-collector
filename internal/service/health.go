// Package service содержит бизнес-логику работы с метриками.
package service

import (
	"context"
)

//go:generate mockgen -destination=mocks/mock_health_repository.go -package=mocks . HealthRepository
type HealthRepository interface {
	Ping(ctx context.Context) error
}

// HealthService checks the health of the service.
type HealthService struct {
	repo HealthRepository
}

// NewHealthService creates a new HealthService.
func NewHealthService(repo HealthRepository) *HealthService {
	return &HealthService{repo: repo}
}

// Ping checks the database connection.
func (svc *HealthService) Ping(ctx context.Context) error {
	return svc.repo.Ping(ctx)
}
