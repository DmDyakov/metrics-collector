package service

import (
	"fmt"
	models "metrics-collector/internal/model"
)

func (svc *MetricsService) UpdateMetricV2(input models.Metrics) (*models.Metrics, error) {
	err := svc.validateMetricFull(&input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	var m models.Metrics

	switch input.MType {
	case models.Counter:
		delta := *input.Delta
		existing, ok := svc.repo.GetMetric(input.ID)

		if ok {
			if existing.MType != models.Counter {
				return nil, ErrMetricTypeMismatch
			}

			if existing.Delta != nil {
				delta += *existing.Delta
			}
		}

		m = models.Metrics{
			ID:    input.ID,
			MType: models.Counter,
			Delta: &delta,
		}

	case models.Gauge:
		m = input

	default:
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, ErrUnknownMetricType)
	}

	updatedMetric := svc.repo.UpdateMetric(m)

	err = svc.validateMetricFull(updatedMetric)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, err)
	}

	return updatedMetric, nil
}

func (svc *MetricsService) GetMetricValueV2(input models.Metrics) (*models.Metrics, error) {
	if err := svc.validateMetricBase(&input); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	m, ok := svc.repo.GetMetric(input.ID)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, ErrMetricNotFound)
	}

	if m.MType != input.MType {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, ErrMetricTypeMismatch)
	}

	err := svc.validateMetricFull(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRepoData, err)
	}

	return m, nil
}
