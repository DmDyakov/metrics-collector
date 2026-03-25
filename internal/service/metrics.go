package service

import (
	"errors"
	"strconv"

	models "metrics-collector/internal/model"
)

type Repository interface {
	GetAllMetrics() map[string]models.Metrics
	GetMetric(metricName string) (*models.Metrics, bool)
	UpdateMetric(metric models.Metrics) *models.Metrics
}

var (
	ErrMetricTypeMismatch  = errors.New("metric type mismatch")
	ErrUnknownMetricType   = errors.New("unknown metric type")
	ErrMetricNotFound      = errors.New("metric not found")
	ErrInvalidCounterValue = errors.New("invalid counter value, should be int")
	ErrInvalidGaugeValue   = errors.New("invalid gauge value, should be float")
)

type MetricsService struct {
	repo Repository
}

func NewMetricsService(repo Repository) *MetricsService {
	return &MetricsService{repo: repo}
}

func (svc *MetricsService) UpdateMetric(metric models.Metrics) (*models.Metrics, error) {
	var m models.Metrics

	switch metric.MType {
	case models.Counter:
		delta := *metric.Delta
		existing, ok := svc.repo.GetMetric(metric.ID)

		if ok {
			if existing.MType != models.Counter {
				return nil, ErrMetricTypeMismatch
			}

			if existing.Delta != nil {
				delta += *existing.Delta
			}
		}

		m = models.Metrics{
			ID:    metric.ID,
			MType: models.Counter,
			Delta: &delta,
		}

	case models.Gauge:
		m = metric

	default:
		return nil, ErrUnknownMetricType
	}

	updatedMetric := svc.repo.UpdateMetric(m)

	return updatedMetric, nil
}

func (svc *MetricsService) GetMetricValue(m models.Metrics) (*models.Metrics, error) {
	metric, ok := svc.repo.GetMetric(m.ID)
	if !ok {
		return nil, ErrMetricNotFound
	}

	if metric.MType != m.MType {
		return nil, ErrMetricTypeMismatch
	}

	switch m.MType {
	case models.Gauge:
		if metric.Value == nil {
			return nil, ErrInvalidGaugeValue
		}
		return metric, nil
	case models.Counter:
		if metric.Delta == nil {
			return nil, ErrInvalidCounterValue
		}
		return metric, nil
	default:
		return nil, ErrUnknownMetricType
	}
}

func (svc *MetricsService) GetAllMetrics() (map[string]string, error) {
	metrics := svc.repo.GetAllMetrics()

	allMetrics := make(map[string]string, len(metrics))

	for name, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				allMetrics[name] = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			}
		case models.Counter:
			if metric.Delta != nil {
				allMetrics[name] = strconv.FormatInt(*metric.Delta, 10)
			}
		}
	}

	return allMetrics, nil

}
