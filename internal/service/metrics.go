package service

import (
	"fmt"
	"metrics-collector/internal/errs"
	models "metrics-collector/internal/model"
	"strconv"
)

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks . Repository
type Repository interface {
	GetAllMetrics() map[string]models.Metrics
	GetMetric(metricName string) (*models.Metrics, bool)
	UpdateMetric(metric models.Metrics) (*models.Metrics, error)
}

type MetricsService struct {
	repo Repository
}

func NewMetricsService(repo Repository) *MetricsService {
	return &MetricsService{repo: repo}
}

func (svc *MetricsService) UpdateMetricByArgs(metricType, metricName, metricValue string) (*models.Metrics, error) {
	input := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case models.Counter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, errs.ErrInvalidCounterValue)
		}

		input.Delta = &delta

		err = svc.validateMetricFull(&input)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
		}

		existing, ok := svc.repo.GetMetric(input.ID)
		if ok {
			if existing.MType != models.Counter {
				return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, errs.ErrMetricTypeMismatch)
			}
			if existing.Delta != nil {
				delta += *existing.Delta
			}
		}

		updated, err := svc.repo.UpdateMetric(models.Metrics{
			ID:    input.ID,
			MType: models.Counter,
			Delta: &delta,
		})

		if err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
		}

		return updated, nil

	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, errs.ErrInvalidGaugeValue)
		}

		input.Value = &value

		updated, err := svc.repo.UpdateMetric(input)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
		}

		return updated, nil

	default:
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, errs.ErrUnknownMetricType)
	}

}

func (svc *MetricsService) GetAllMetrics() ([]models.Metrics, error) {
	metricsMap := svc.repo.GetAllMetrics()

	metrics := make([]models.Metrics, 0, len(metricsMap))

	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (svc *MetricsService) GetMetricValue(metricType, metricName string) (*string, error) {
	input := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	if err := svc.validateMetricBase(&input); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
	}

	m, ok := svc.repo.GetMetric(input.ID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errs.ErrMetricNotFound, input.ID)
	}

	if m.MType != input.MType {
		return nil, fmt.Errorf("%w: expected %s for id: %s, received %s",
			errs.ErrMetricTypeMismatch,
			m.MType,
			input.ID,
			input.MType)
	}

	value, err := formatToString(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
	}

	return &value, nil
}

func formatToString(m *models.Metrics) (string, error) {
	switch m.MType {
	case models.Counter:
		return strconv.FormatInt(*m.Delta, 10), nil
	case models.Gauge:
		return strconv.FormatFloat(*m.Value, 'f', -1, 64), nil
	default:
		return "", errs.ErrUnknownMetricType
	}
}

// New API (JSON-based)
func (svc *MetricsService) UpdateMetricByJSON(input models.Metrics) (*models.Metrics, error) {
	err := svc.validateMetricFull(&input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
	}

	var m models.Metrics

	switch input.MType {
	case models.Counter:
		delta := *input.Delta
		existing, ok := svc.repo.GetMetric(input.ID)

		if ok {
			if existing.MType != models.Counter {
				return nil, fmt.Errorf("%w: expected %s for id: %s, received %s",
					errs.ErrMetricTypeMismatch,
					existing.MType, input.ID,
					input.MType)
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
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, errs.ErrUnknownMetricType)
	}

	updatedMetric, err := svc.repo.UpdateMetric(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidResponse, err)
	}

	return updatedMetric, nil
}

func (svc *MetricsService) GetMetric(input models.Metrics) (*models.Metrics, error) {
	if err := svc.validateMetricBase(&input); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
	}

	m, ok := svc.repo.GetMetric(input.ID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errs.ErrMetricNotFound, input.ID)
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
