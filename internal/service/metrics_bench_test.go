package service

import (
	"fmt"
	"testing"

	models "metrics-collector/internal/model"
)

func BenchmarkDeduplicateBatch(b *testing.B) {
	svc := NewMetricsService(nil) // store не нужен для deduplicateBatch
	batch := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		v := float64(i % 10)
		batch[i] = models.Metrics{
			ID:    fmt.Sprintf("m_%d", i%10),
			MType: "gauge",
			Value: &v,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.deduplicateBatch(batch)
	}
}

func BenchmarkValidateMetricFull(b *testing.B) {
	svc := NewMetricsService(nil)
	v := 42.5
	m := models.Metrics{ID: "test", MType: "gauge", Value: &v}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.validateMetricFull(&m)
	}
}
