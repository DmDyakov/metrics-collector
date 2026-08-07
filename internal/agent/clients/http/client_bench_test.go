package client

import (
	"metrics-collector/internal/domain/metrics"
	"testing"
)

func BenchmarkToDto(b *testing.B) {
	c := &Client{}
	metrics := map[string]float64{
		"cpu": 42.5, "memory": 1024, "disk": 256,
		metrics.PollCount: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.toDto(metrics)
	}
}
