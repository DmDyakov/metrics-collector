package service

import (
	"fmt"
	"metrics-collector/internal/domain/metrics"
	"metrics-collector/internal/server/errs"
	models "metrics-collector/internal/server/model"
)

func (svc *MetricsService) validateRequired(m models.Metrics) error {
	if m.MType == "" {
		return errs.ErrMetricTypeRequired
	}
	if m.ID == "" {
		return errs.ErrMetricNameRequired
	}
	return nil
}

func (svc *MetricsService) validateMetricType(m models.Metrics) error {
	if m.MType != metrics.Gauge && m.MType != metrics.Counter {
		return fmt.Errorf("%w: %s", errs.ErrUnknownMetricType, m.MType)
	}
	return nil
}

func (svc *MetricsService) validateCounterDeltaRequired(m models.Metrics) error {
	if m.MType == metrics.Counter && m.Delta == nil {
		return errs.ErrMetricDeltaForCountRequired
	}
	return nil
}

func (svc *MetricsService) validateGaugeValueRequired(m models.Metrics) error {
	if m.MType == metrics.Gauge && m.Value == nil {
		return errs.ErrMetricValueForGaugeRequired
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
	case metrics.Gauge:
		return svc.validateGaugeValueRequired(*m)
	case metrics.Counter:
		return svc.validateCounterDeltaRequired(*m)
	default:
		return errs.ErrUnknownMetricType
	}
}
