package service

import (
	"errors"
	"strconv"

	models "metrics-collector/internal/model"
)

type Repository interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
}

type MetricsService struct {
	repo Repository
}

func NewMetricsService(repo Repository) *MetricsService {
	return &MetricsService{repo: repo}
}

func (svc *MetricsService) Update(metricType, metricName, metricValueRaw string) error {
	switch metricType {
	case models.Counter:
		metricValue, err := strconv.ParseInt(metricValueRaw, 10, 64)
		if err != nil {
			return err
		}

		svc.repo.UpdateCounter(metricName, metricValue)

	case models.Gauge:
		metricValue, err := strconv.ParseFloat(metricValueRaw, 64)
		if err != nil {
			return err
		}

		svc.repo.UpdateGauge(metricName, metricValue)

	default:
		return errors.New("unknown metric type")
	}

	return nil
}
