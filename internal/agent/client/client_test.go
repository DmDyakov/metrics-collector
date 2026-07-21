package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"syscall"
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

func TestClient_IsRetriable(t *testing.T) {
	t.Run("timeout error is retriable", func(t *testing.T) {
		err := &net.DNSError{IsTimeout: true}
		assert.True(t, isRetriable(nil, err))
	})

	t.Run("connection refused is retriable", func(t *testing.T) {
		assert.True(t, isRetriable(nil, syscall.ECONNREFUSED))
	})

	t.Run("status 429 is retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusTooManyRequests}
		assert.True(t, isRetriable(resp, nil))
	})

	t.Run("status 200 is not retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusOK}
		assert.False(t, isRetriable(resp, nil))
	})

	t.Run("random error is not retriable", func(t *testing.T) {
		assert.False(t, isRetriable(nil, errors.New("random error")))
	})
}

func TestClient_CreateSignature(t *testing.T) {
	t.Run("creates HMAC signature", func(t *testing.T) {
		c := New("localhost:8080", "secret", zap.NewNop(), compress.NewGzip())

		sig := c.createSignature([]byte("test"))
		assert.NotEmpty(t, sig)
		assert.Len(t, sig, 32)
	})
}
