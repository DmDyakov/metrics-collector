package mem

import (
	"fmt"
	"testing"

	models "metrics-collector/internal/server/model"
)

func BenchmarkSaveMetric(b *testing.B) {
	s := NewMemStorage()
	v := 42.5
	m := models.Metrics{ID: "test", MType: "gauge", Value: &v}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SaveMetric(m)
	}
}

func BenchmarkSaveMetric_Parallel(b *testing.B) {
	s := NewMemStorage()
	v := 42.5
	m := models.Metrics{ID: "test", MType: "gauge", Value: &v}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.SaveMetric(m)
		}
	})
}

func BenchmarkGetAll(b *testing.B) {
	s := NewMemStorage()
	for i := 0; i < 100; i++ {
		v := float64(i)
		s.SaveMetric(models.Metrics{
			ID:    fmt.Sprintf("metric_%d", i),
			MType: "gauge",
			Value: &v,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.GetAll()
	}
}

func BenchmarkSaveBatch(b *testing.B) {
	s := NewMemStorage()
	batch := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		v := float64(i)
		batch[i] = models.Metrics{
			ID:    fmt.Sprintf("metric_%d", i),
			MType: "gauge",
			Value: &v,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SaveBatch(batch)
	}
}
