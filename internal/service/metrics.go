package service

import (
	models "metrics-collector/internal/model"
)

type Repository interface {
	GetAllMetrics() map[string]models.Metrics
	GetMetric(metricName string) (*models.Metrics, bool)
	UpdateMetric(metric models.Metrics) *models.Metrics
}

type MetricsService struct {
	repo Repository
}

func NewMetricsService(repo Repository) *MetricsService {
	return &MetricsService{repo: repo}
}
