package repository

import models "metrics-collector/internal/model"

type MemStorage struct {
	metrics map[string]models.Metrics
}

func (ms MemStorage) GetAllMetrics() map[string]models.Metrics {
	metrics := make(map[string]models.Metrics)
	for metricName, metric := range ms.metrics {
		metrics[metricName] = copyMetric(metric)
	}
	return metrics
}

func (ms *MemStorage) GetMetric(metricName string) (models.Metrics, bool) {
	metric, ok := ms.metrics[metricName]
	return copyMetric(metric), ok
}

func (ms *MemStorage) UpdateMetric(metric models.Metrics) {
	ms.metrics[metric.ID] = metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func copyMetric(m models.Metrics) models.Metrics {
	copy := models.Metrics{
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
