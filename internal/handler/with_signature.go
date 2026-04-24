package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type signResponseWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (w *signResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *signResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = statusCode
}

func (h *Handler) WithSignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.secretKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		if err := h.checkRequestSignature(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		srw := &signResponseWriter{
			ResponseWriter: w,
			buffer:         &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(srw, r)

		if srw.buffer.Len() > 0 {
			signature := h.createSignature(srw.buffer.Bytes())
			w.Header().Set("HashSHA256", hex.EncodeToString(signature))
		}

		w.WriteHeader(srw.statusCode)

		if r.Method == http.MethodHead {
			return
		}

		if srw.buffer.Len() > 0 {
			w.Write(srw.buffer.Bytes())
		}
	})
}

func (h *Handler) checkRequestSignature(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("body_size", len(body)))
		return errors.New("failed to read request body")
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body)) // восстанавливаем тело

	received, err := hex.DecodeString(r.Header.Get("HashSHA256"))
	if err != nil {
		return errors.New("invalid signature format")
	}

	expected := h.createSignature(body)

	if !hmac.Equal(received, expected) {
		h.logger.Error("signed body mismatch",
			zap.String("expected", hex.EncodeToString(expected)),
			zap.String("received", hex.EncodeToString(received)),
		)
		return errors.New("invalid request signature")
	}

	return nil
}

func (h *Handler) createSignature(data []byte) []byte {
	hmacHash := hmac.New(sha256.New, []byte(h.secretKey))
	hmacHash.Write(data)

	return hmacHash.Sum(nil)
}
