package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"syscall"
	"testing"

	models "metrics-collector/internal/model"
	"metrics-collector/pkg/compress"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient_ToDto(t *testing.T) {
	c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())

	t.Run("converts metrics to DTO", func(t *testing.T) {
		metrics := map[string]float64{
			"cpu":            42.5,
			models.PollCount: 10,
		}

		dto := c.toDto(metrics)
		assert.Len(t, dto, 2)
	})

	t.Run("empty metrics returns empty slice", func(t *testing.T) {
		dto := c.toDto(map[string]float64{})
		assert.Empty(t, dto)
	})
}

func TestClient_SendMetrics(t *testing.T) {
	c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())

	t.Run("skips empty batch", func(t *testing.T) {
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
	s, _ := signer.New("secret")

	t.Run("creates HMAC signature", func(t *testing.T) {
		sig := s.CreateSignature([]byte("test"))
		assert.NotEmpty(t, sig)
		assert.Len(t, sig, 32)
	})
}

func TestClient_Encryption(t *testing.T) {
	t.Run("encrypts and decrypts data", func(t *testing.T) {
		enc, err := encryptor.New("testdata/public.pem", "testdata/private.pem")
		if err != nil {
			t.Skip("test keys not found, skipping encryption test")
		}

		data := []byte("test data")
		encrypted, err := enc.Encrypt(data)
		assert.NoError(t, err)
		assert.NotEqual(t, data, encrypted)

		decrypted, err := enc.Decrypt(encrypted)
		assert.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})
}
