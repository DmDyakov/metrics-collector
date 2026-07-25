package middleware

import (
	"bytes"
	"encoding/hex"
	"errors"
	"metrics-collector/pkg/signer"
	"net/http"

	"go.uber.org/zap"
)

type signResponseWriter struct {
	http.ResponseWriter
	buffer      *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func (w *signResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.buffer.Write(data)
}

func (w *signResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.statusCode = statusCode
}

// WithSignature middleware verifies the HMAC signature of incoming requests
// and signs outgoing responses.
func WithSignature(logger *zap.Logger, s *signer.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if s == nil {
				next.ServeHTTP(w, r)
				return
			}

			if err := s.CheckRequestSignature(r); err != nil {
				if errors.Is(err, signer.ErrInvalidSignature) {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				logger.Error("failed WithSignature middleware",
					zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			srw := &signResponseWriter{
				ResponseWriter: w,
				buffer:         &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(srw, r)

			if srw.buffer.Len() > 0 {
				signature := s.CreateSignature(srw.buffer.Bytes())
				w.Header().Set("HashSHA256", hex.EncodeToString(signature))
			}

			w.WriteHeader(srw.statusCode)

			if r.Method != http.MethodHead && srw.buffer.Len() > 0 {
				if _, err := w.Write(srw.buffer.Bytes()); err != nil {
					logger.Error("failed to write response body",
						zap.Error(err))
				}
			}
		})
	}
}
