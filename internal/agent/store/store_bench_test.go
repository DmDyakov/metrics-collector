package store

import (
	"testing"
)

func BenchmarkUpdateMetrics(b *testing.B) {
	s := New()
	metrics := map[string]float64{
		"cpu": 42.5, "memory": 1024, "disk": 256,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.UpdateMetrics(metrics)
	}
}

func BenchmarkGetMetricsSnapshot(b *testing.B) {
	s := New()
	s.UpdateMetrics(map[string]float64{
		"cpu": 42.5, "memory": 1024,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.GetMetricsSnapshot()
	}
}
