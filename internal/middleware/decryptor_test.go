package middleware

import (
	"bytes"
	"io"
	"metrics-collector/pkg/encryptor"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWithDecryption(t *testing.T) {
	logger := zap.NewNop()

	t.Run("nil encryptor passes through", func(t *testing.T) {
		handler := WithDecryption(logger, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("test")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("decrypts request body", func(t *testing.T) {
		enc, err := encryptor.New("testdata/public.pem", "testdata/private.pem")
		if err != nil {
			t.Skip("test keys not found")
		}

		original := []byte("secret data")
		encrypted, err := enc.Encrypt(original)
		require.NoError(t, err)

		var received []byte
		handler := WithDecryption(logger, enc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(encrypted))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, original, received)
	})

	t.Run("returns 400 on invalid encrypted data", func(t *testing.T) {
		enc, err := encryptor.New("", "testdata/private.pem")
		if err != nil {
			t.Skip("test keys not found")
		}

		handler := WithDecryption(logger, enc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("garbage")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
