package service

import (
	"fmt"
	"metrics-collector/internal/errs"
	models "metrics-collector/internal/model"
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
	if m.MType != models.Gauge && m.MType != models.Counter {
		return fmt.Errorf("%w: %s", errs.ErrUnknownMetricType, m.MType)
	}
	return nil
}

func (svc *MetricsService) validateCounterDeltaRequired(m models.Metrics) error {
	if m.MType == models.Counter && m.Delta == nil {
		return errs.ErrMetricDeltaForCountRequired
	}
	return nil
}

func (svc *MetricsService) validateGaugeValueRequired(m models.Metrics) error {
	if m.MType == models.Gauge && m.Value == nil {
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
	case models.Gauge:
		return svc.validateGaugeValueRequired(*m)
	case models.Counter:
		return svc.validateCounterDeltaRequired(*m)
	default:
		return errs.ErrUnknownMetricType
	}
}
