package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"metrics-collector/internal/errs"
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

func checkRequestSignature(r *http.Request, secretKey string) error {
	if r.Header.Get("HashSHA256") == "" {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return &errs.ErrRequestBodyRead{
			Method:   r.Method,
			URL:      r.URL.String(),
			BodySize: len(body),
			Err:      err,
		}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body)) // восстанавливаем тело

	headerValue := r.Header.Get("HashSHA256")
	received, err := hex.DecodeString(headerValue)
	if err != nil {
		return &errs.ErrInvalidHexFormat{
			HeaderValue: headerValue,
			Err:         err,
		}
	}

	expected := createSignature(body, secretKey)

	if !hmac.Equal(received, expected) {
		return &errs.ErrSignedBodyMismatch{
			Expected: hex.EncodeToString(expected),
			Received: hex.EncodeToString(received),
		}
	}

	return nil
}

func createSignature(data []byte, secretKey string) []byte {
	hmacHash := hmac.New(sha256.New, []byte(secretKey))
	hmacHash.Write(data)

	return hmacHash.Sum(nil)
}

func WithSignature(logger *zap.Logger, secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secretKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if err := checkRequestSignature(r, secretKey); err != nil {
				if errors.Is(err, errs.ErrInvalidSignature) {
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
				signature := createSignature(srw.buffer.Bytes(), secretKey)
				w.Header().Set("HashSHA256", hex.EncodeToString(signature))
			}

			w.WriteHeader(srw.statusCode)

			if r.Method != http.MethodHead && srw.buffer.Len() > 0 {
				w.Write(srw.buffer.Bytes())
			}
		})
	}
}
