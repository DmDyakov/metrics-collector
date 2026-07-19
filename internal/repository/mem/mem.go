// Package mem реализует хранение метрик в памяти.
package mem

import (
	models "metrics-collector/internal/model"
	"sync"
)

type MemStorage struct {
	mu      sync.RWMutex
	metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (ms *MemStorage) GetAll() map[string]models.Metrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make(map[string]models.Metrics, len(ms.metrics))
	for k, v := range ms.metrics {
		result[k] = v
	}
	return result
}

func (ms *MemStorage) GetMetricByName(metricName string) (*models.Metrics, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	m, ok := ms.metrics[metricName]
	return &m, ok
}

func (ms *MemStorage) SaveMetric(metric models.Metrics) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.metrics[metric.ID] = metric
}

func (ms *MemStorage) SaveBatch(metrics []models.Metrics) *int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, m := range metrics {
		ms.metrics[m.ID] = m
	}
	count := len(metrics)
	return &count
}
