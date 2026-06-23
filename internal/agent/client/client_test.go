package client

import (
	"context"
	"testing"

	"metrics-collector/internal/agent/compress"
	models "metrics-collector/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient_ToDto(t *testing.T) {
	t.Run("converts metrics to DTO", func(t *testing.T) {
		c := New("localhost:8080", "", zap.NewNop(), compress.NewGzip())

		metrics := map[string]float64{
			"cpu":            42.5,
			models.PollCount: 10,
		}

		dto := c.toDto(metrics)
		assert.Len(t, dto, 2)
	})

	t.Run("empty metrics returns empty slice", func(t *testing.T) {
		c := New("localhost:8080", "", zap.NewNop(), compress.NewGzip())
		dto := c.toDto(map[string]float64{})
		assert.Empty(t, dto)
	})
}

func TestClient_SendMetrics(t *testing.T) {
	t.Run("skips empty batch", func(t *testing.T) {
		c := New("localhost:8080", "", zap.NewNop(), compress.NewGzip())
		err := c.SendMetrics(context.Background(), map[string]float64{})
		assert.NoError(t, err)
	})
}
