package service

import (
	"errors"
	"fmt"
	models "metrics-collector/internal/model"
)

var (
	ErrInvalidRepoData = errors.New("invalid metric data in repo")
	ErrInvalidRequest  = errors.New("invalid request data")

	ErrMetricTypeRequired          = errors.New("metric type is required")
	ErrMetricNameRequired          = errors.New("metric name is required")
	ErrMetricValueForGaugeRequired = errors.New("metric value is required for gauge")
	ErrMetricDeltaForCountRequired = errors.New("metric delta is required for counter")
	ErrMetricTypeMismatch          = errors.New("metric type mismatch")
	ErrUnknownMetricType           = errors.New("unknown metric type")
	ErrMetricNotFound              = errors.New("metric not found")
	ErrInvalidCounterValue         = errors.New("invalid counter value, should be int")
	ErrInvalidGaugeValue           = errors.New("invalid gauge value, should be float")
)

func (svc *MetricsService) validateRequired(m models.Metrics) error {
	if m.MType == "" {
		return ErrMetricTypeRequired
	}
	if m.ID == "" {
		return ErrMetricNameRequired
	}
	return nil
}

func (svc *MetricsService) validateMetricType(m models.Metrics) error {
	if m.MType != models.Gauge && m.MType != models.Counter {
		return fmt.Errorf("%w: %s", ErrUnknownMetricType, m.MType)
	}
	return nil
}

func (svc *MetricsService) validateCounterDeltaRequired(m models.Metrics) error {
	if m.MType == models.Counter && m.Delta == nil {
		return ErrMetricDeltaForCountRequired
	}
	return nil
}

func (svc *MetricsService) validateGaugeValueRequired(m models.Metrics) error {
	if m.MType == models.Gauge && m.Value == nil {
		return ErrMetricValueForGaugeRequired
	}
	return nil
}

func (svc *MetricsService) validateMetricBase(m *models.Metrics) error {
	if err := svc.validateRequired(*m); err != nil {
		return err
	}

	if err := svc.validateMetricType(*m); err != nil {
		return err
	}

	return nil
}

func (svc *MetricsService) validateMetricFull(m *models.Metrics) error {
	if err := svc.validateRequired(*m); err != nil {
		return err
	}

	if err := svc.validateMetricType(*m); err != nil {
		return err
	}

	switch m.MType {
	case models.Gauge:
		return svc.validateGaugeValueRequired(*m)
	case models.Counter:
		return svc.validateCounterDeltaRequired(*m)
	default:
		return ErrUnknownMetricType
	}
}
