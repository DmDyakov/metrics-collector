package repository

import (
	models "metrics-collector/internal/model"
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (ms *MemStorage) GetAllMetrics() map[string]models.Metrics {
	metrics := make(map[string]models.Metrics)
	for metricName, metric := range ms.metrics {
		copied := copyMetric(metric)
		metrics[metricName] = *copied
	}
	return metrics
}

func (ms *MemStorage) GetMetric(metricName string) (*models.Metrics, bool) {
	metric, ok := ms.metrics[metricName]
	if !ok {
		return nil, false
	}
	return copyMetric(metric), true
}

func (ms *MemStorage) UpdateMetric(metric models.Metrics) *models.Metrics {
	ms.metrics[metric.ID] = metric

	result := copyMetric(ms.metrics[metric.ID])

	return result
}

func (ms *MemStorage) replaceMetrics(metrics map[string]models.Metrics) {
	ms.metrics = metrics
}

func copyMetric(m models.Metrics) *models.Metrics {
	copy := &models.Metrics{
		ID:    m.ID,
		MType: m.MType,
		Hash:  m.Hash,
	}

	if m.Delta != nil {
		delta := *m.Delta
		copy.Delta = &delta
	}

	if m.Value != nil {
		val := *m.Value
		copy.Value = &val
	}

	return copy
}
