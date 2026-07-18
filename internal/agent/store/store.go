package store

import (
	"sync"
)

type Store struct {
	metrics map[string]float64
	mu      sync.RWMutex
}

func New() *Store {
	return &Store{
		metrics: make(map[string]float64),
	}
}

func (s *Store) UpdateMetrics(metrics map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range metrics {
		s.metrics[key] = value
	}
}

func (s *Store) GetMetricsSnapshot() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := make(map[string]float64, len(s.metrics))
	for k, v := range s.metrics {
		copy[k] = v
	}
	return copy
}
