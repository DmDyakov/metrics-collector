package service

import (
	"errors"
	"strconv"

	models "metrics-collector/internal/model"
)

type Repository interface {
	GetAllMetrics() map[string]models.Metrics
	GetMetric(metricName string) (models.Metrics, bool)
	UpdateMetric(metric models.Metrics)
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

func (svc *MetricsService) UpdateMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case models.Counter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrInvalidCounterValue
		}

		existing, ok := svc.repo.GetMetric(metricName)

		if ok {
			if existing.MType != models.Counter {
				return ErrMetricTypeMismatch
			}

			if existing.Delta != nil {
				delta += *existing.Delta
			}
		}

		metric := models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &delta,
		}

		svc.repo.UpdateMetric(metric)

	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrInvalidGaugeValue
		}

		metric := models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		}

		svc.repo.UpdateMetric(metric)

	default:
		return ErrUnknownMetricType
	}

	return nil
}

func (svc *MetricsService) GetMetricValue(metricType, metricName string) (string, error) {
	metric, ok := svc.repo.GetMetric(metricName)
	if !ok {
		return "", ErrMetricNotFound
	}

	if metric.MType != metricType {
		return "", ErrMetricTypeMismatch
	}

	switch metricType {
	case models.Gauge:
		if metric.Value == nil {
			return "", ErrInvalidGaugeValue
		}
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
	case models.Counter:
		if metric.Delta == nil {
			return "", ErrInvalidCounterValue
		}
		return strconv.FormatInt(*metric.Delta, 10), nil
	default:
		return "", ErrUnknownMetricType
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
