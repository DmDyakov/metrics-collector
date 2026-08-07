// Package dto содержит объекты передачи данных между компонентами системы.
package dto

import "metrics-collector/internal/domain/metrics"

// Metric представляет собой DTO метрики
//
// generate:reset
type Metric struct {
	ID    string       `json:"id"`
	Type  metrics.Type `json:"type"`
	Delta *int64       `json:"delta,omitempty"` // для Counter
	Value *float64     `json:"value,omitempty"` // для Gauge
}
