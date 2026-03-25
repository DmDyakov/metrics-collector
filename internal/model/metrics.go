package models

import (
	"errors"
	"fmt"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrMetricTypeRequired          = errors.New("metric type is required")
	ErrMetricNameRequired          = errors.New("metric name is required")
	ErrMetricValueForGaugeRequired = errors.New("metric value is required for gauge")
	ErrMetricDeltaForCountRequired = errors.New("metric delta is required for counter")
	ErrUnknownMetricType           = errors.New("unknown metric type")
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) ValidateBaseRequired() error {
	if m.MType == "" {
		return ErrMetricTypeRequired
	}
	if m.ID == "" {
		return ErrMetricNameRequired
	}
	return nil
}

func (m *Metrics) ValidateMetricType() error {
	if m.MType != Gauge && m.MType != Counter {
		return fmt.Errorf("%w: %s", ErrUnknownMetricType, m.MType)
	}
	return nil
}

func (m *Metrics) ValidateCounterDeltaRequired() error {
	if m.MType == Counter && m.Delta == nil {
		return ErrMetricDeltaForCountRequired
	}
	return nil
}

func (m *Metrics) ValidateGaugeValueRequired() error {
	if m.MType == Gauge && m.Value == nil {
		return ErrMetricValueForGaugeRequired
	}
	return nil
}

func (m *Metrics) ValidateBase() error {
	if err := m.ValidateBaseRequired(); err != nil {
		return err
	}

	if err := m.ValidateMetricType(); err != nil {
		return err
	}

	return nil
}

func (m *Metrics) ValidateForUpdate() error {
	if err := m.ValidateBaseRequired(); err != nil {
		return err
	}

	if err := m.ValidateMetricType(); err != nil {
		return err
	}

	if err := m.ValidateCounterDeltaRequired(); err != nil {
		return err
	}

	if err := m.ValidateGaugeValueRequired(); err != nil {
		return err
	}

	return nil
}
