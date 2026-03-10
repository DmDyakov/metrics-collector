package service

import (
	"errors"
	"strconv"

	models "metrics-collector/internal/model"
)

type Repository interface {
	GetAllMetricsRaw() (gauges map[string]float64, counters map[string]int64)
	UpdateCounterMetric(name string, value int64)
	UpdateGaugeMetric(name string, value float64)
	GetGaugeMetricValue(name string) (float64, bool)
	GetCountMetricValue(name string) (int64, bool)
}

var (
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

func (svc *MetricsService) UpdateMetric(metricType, metricName, metricValueRaw string) error {
	switch metricType {
	case models.Counter:
		metricValue, err := strconv.ParseInt(metricValueRaw, 10, 64)
		if err != nil {
			return ErrInvalidCounterValue
		}

		svc.repo.UpdateCounterMetric(metricName, metricValue)

	case models.Gauge:
		metricValue, err := strconv.ParseFloat(metricValueRaw, 64)
		if err != nil {
			return ErrInvalidGaugeValue
		}

		svc.repo.UpdateGaugeMetric(metricName, metricValue)

	default:
		return ErrUnknownMetricType
	}

	return nil
}

func (svc *MetricsService) GetMetricValue(metricType, metricName string) (string, error) {
	switch metricType {
	case "gauge":
		value, ok := svc.repo.GetGaugeMetricValue(metricName)
		if !ok {
			return "", ErrMetricNotFound
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil

	case "count":
		value, ok := svc.repo.GetCountMetricValue(metricName)
		if !ok {
			return "", ErrMetricNotFound
		}
		return strconv.FormatInt(value, 10), nil

	default:
		return "", ErrUnknownMetricType
	}
}

func (svc *MetricsService) GetAllMetrics() (map[string]string, error) {
	gauges, counters := svc.repo.GetAllMetricsRaw()

	allMetrics := make(map[string]string)

	for name, value := range gauges {
		allMetrics[name] = strconv.FormatFloat(value, 'f', -1, 64)
	}

	for name, value := range counters {
		allMetrics[name] = strconv.FormatInt(value, 10)
	}

	return allMetrics, nil

}
