package service

import (
	"context"
	"fmt"
	"metrics-collector/internal/domain/metrics"
	"metrics-collector/internal/server/errs"
	models "metrics-collector/internal/server/model"
)

//go:generate mockgen -destination=mocks/mock_metrics_repository.go -package=mocks . MetricsRepository
type MetricsRepository interface {
	GetAllMetrics() map[string]models.Metrics
	GetMetric(metricName string) (*models.Metrics, bool)
	SaveMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error)
	SaveMetricsBatch(ctx context.Context, metrics []models.Metrics) (*int, error)
}

// MetricsService contains the business logic for metrics operations.
type MetricsService struct {
	repo MetricsRepository
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(repo MetricsRepository) *MetricsService {
	return &MetricsService{repo: repo}
}

// GetAllMetrics returns all stored metrics.
func (svc *MetricsService) GetAllMetrics() ([]models.Metrics, error) {
	metricsMap := svc.repo.GetAllMetrics()

	metrics := make([]models.Metrics, 0, len(metricsMap))

	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// UpdateMetric updates or creates a metric.
func (svc *MetricsService) UpdateMetric(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	if err := svc.validateMetricFull(&m); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
	}

	if m.MType == metrics.Counter {
		existing, ok := svc.repo.GetMetric(m.ID)
		if ok {
			if existing.MType != metrics.Counter {
				return nil, fmt.Errorf("%w: expected %s for id: %s, received %s",
					errs.ErrMetricTypeMismatch, existing.MType, m.ID, m.MType)
			}
			if existing.Delta != nil {
				sum := *m.Delta + *existing.Delta
				m.Delta = &sum
			}
		}
	}

	updated, err := svc.repo.SaveMetric(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
	}

	return updated, nil
}

// UpdateMetrics batch-updates metrics.
func (svc *MetricsService) UpdateMetrics(ctx context.Context, batch []models.Metrics) (*int, error) {
	for _, input := range batch {
		if err := svc.validateMetricFull(&input); err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
		}
	}

	deduped := svc.deduplicateBatch(batch)

	for idx, input := range deduped {
		if input.MType == metrics.Counter {
			existing, ok := svc.repo.GetMetric(input.ID)
			if !ok {
				continue
			}

			if existing.MType != metrics.Counter {
				return nil, fmt.Errorf("%w: expected %s for id: %s, received %s",
					errs.ErrMetricTypeMismatch,
					existing.MType, input.ID,
					input.MType)
			}

			sum := *input.Delta + *existing.Delta
			deduped[idx].Delta = &sum
		}
	}

	count, err := svc.repo.SaveMetricsBatch(ctx, deduped)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
	}

	return count, nil
}

// GetMetric returns a metric.
func (svc *MetricsService) GetMetric(input models.Metrics) (*models.Metrics, error) {
	if err := svc.validateMetricBase(&input); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
	}

	m, ok := svc.repo.GetMetric(input.ID)
	if !ok {
		return nil, &errs.MetricNotFoundError{Type: input.MType, Name: input.ID}
	}

	if m.MType != input.MType {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, errs.ErrMetricTypeMismatch)
	}

	err := svc.validateMetricFull(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
	}

	return m, nil
}

func (svc *MetricsService) deduplicateBatch(batch []models.Metrics) []models.Metrics {
	aggregated := make(map[string]*models.Metrics)

	for _, m := range batch {
		key := m.ID + ":" + string(m.MType)

		if existing, ok := aggregated[key]; ok {
			if m.MType == metrics.Counter {
				sum := *existing.Delta + *m.Delta
				existing.Delta = &sum
			} else {
				existing.Value = m.Value
			}
		} else {
			copied := m
			aggregated[key] = &copied
		}
	}

	result := make([]models.Metrics, 0, len(aggregated))
	for _, m := range aggregated {
		result = append(result, *m)
	}
	return result
}
