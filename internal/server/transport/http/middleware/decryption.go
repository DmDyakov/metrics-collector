package middleware

import (
	"bytes"
	"io"
	"metrics-collector/pkg/encryptor"
	"net/http"

	"go.uber.org/zap"
)

// WithDecryption расшифровывает тело входящего запроса, если задан encryptor.
func WithDecryption(logger *zap.Logger, enc *encryptor.Encryptor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enc == nil {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read request body for decryption", zap.Error(err))
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			decrypted, err := enc.Decrypt(body)
			if err != nil {
				logger.Error("failed to decrypt request body", zap.Error(err))
				http.Error(w, "Bad request: decryption failed", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))

			next.ServeHTTP(w, r)
		})
	}
}
