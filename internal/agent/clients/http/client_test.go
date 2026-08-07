package client

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	models "metrics-collector/internal/server/model"
	"metrics-collector/pkg/compress"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Run("skips empty batch", func(t *testing.T) {
		c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		err := c.SendMetrics(context.Background(), map[string]float64{})
		assert.NoError(t, err)
	})

	t.Run("sends metrics to server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/updates", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
			assert.Equal(t, "100.100.100.100", r.Header.Get("X-Real-IP"))

			body, _ := io.ReadAll(r.Body)
			assert.NotEmpty(t, body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := New(server.Listener.Addr().String(), "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())

		metrics := map[string]float64{"cpu": 42.5}
		err := c.SendMetrics(context.Background(), metrics)
		assert.NoError(t, err)
	})

	t.Run("retries on server error", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := New(server.Listener.Addr().String(), "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		c.httpClient.Timeout = 1 * time.Second

		metrics := map[string]float64{"cpu": 42.5}
		err := c.SendMetrics(context.Background(), metrics)
		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("stops retry on canceled context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		c := New(server.Listener.Addr().String(), "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		c.httpClient.Timeout = 1 * time.Second

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		metrics := map[string]float64{"cpu": 42.5}
		err := c.SendMetrics(ctx, metrics)
		assert.Error(t, err)
	})

	t.Run("sends encrypted and signed metrics", func(t *testing.T) {
		enc, err := encryptor.New("testdata/public.pem", "testdata/private.pem")
		if err != nil {
			t.Skip("test keys not found")
		}
		s, err := signer.New("secret")
		require.NoError(t, err)

		var receivedSig string
		var receivedBody []byte

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSig = r.Header.Get("HashSHA256")
			receivedBody, _ = io.ReadAll(r.Body)

			sig := s.CreateSignature([]byte(`{"status":"ok"}`))
			w.Header().Set("HashSHA256", hex.EncodeToString(sig))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		c := New(server.Listener.Addr().String(), "100.100.100.100", zap.NewNop(), s, enc, compress.NewGzip())

		metrics := map[string]float64{"cpu": 42.5}
		err = c.SendMetrics(context.Background(), metrics)
		assert.NoError(t, err)
		assert.NotEmpty(t, receivedSig)
		assert.NotEmpty(t, receivedBody)
	})
}

func TestClient_WithRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		attempts := 0
		doRequest := func() (*http.Response, error) {
			attempts++
			return &http.Response{StatusCode: http.StatusOK}, nil
		}

		err := c.withRetry(context.Background(), doRequest)
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("success after retries", func(t *testing.T) {
		c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		attempts := 0
		doRequest := func() (*http.Response, error) {
			attempts++
			if attempts < 2 {
				return &http.Response{StatusCode: http.StatusServiceUnavailable}, nil
			}
			return &http.Response{StatusCode: http.StatusOK}, nil
		}

		err := c.withRetry(context.Background(), doRequest)
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("returns error after max retries", func(t *testing.T) {
		c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		doRequest := func() (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusServiceUnavailable}, nil
		}

		err := c.withRetry(context.Background(), doRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status")
	})

	t.Run("returns error on non-retriable status", func(t *testing.T) {
		c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		doRequest := func() (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusBadRequest}, nil
		}

		err := c.withRetry(context.Background(), doRequest)
		assert.Error(t, err)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		c := New("localhost:8080", "100.100.100.100", zap.NewNop(), nil, nil, compress.NewGzip())
		ctx, cancel := context.WithCancel(context.Background())

		doRequest := func() (*http.Response, error) {
			cancel() // отменяем после первого вызова
			return &http.Response{StatusCode: http.StatusServiceUnavailable}, nil
		}

		err := c.withRetry(ctx, doRequest)
		assert.Error(t, err)
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

	t.Run("connection reset is retriable", func(t *testing.T) {
		assert.True(t, isRetriable(nil, syscall.ECONNRESET))
	})

	t.Run("status 429 is retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusTooManyRequests}
		assert.True(t, isRetriable(resp, nil))
	})

	t.Run("status 502 is retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusBadGateway}
		assert.True(t, isRetriable(resp, nil))
	})

	t.Run("status 503 is retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusServiceUnavailable}
		assert.True(t, isRetriable(resp, nil))
	})

	t.Run("status 504 is retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusGatewayTimeout}
		assert.True(t, isRetriable(resp, nil))
	})

	t.Run("status 408 is retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusRequestTimeout}
		assert.True(t, isRetriable(resp, nil))
	})

	t.Run("status 200 is not retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusOK}
		assert.False(t, isRetriable(resp, nil))
	})

	t.Run("status 400 is not retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusBadRequest}
		assert.False(t, isRetriable(resp, nil))
	})

	t.Run("status 500 is not retriable", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusInternalServerError}
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
