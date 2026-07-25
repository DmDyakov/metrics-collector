// Package middleware содержит middleware для HTTP-сервера
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"metrics-collector/internal/audit"
	models "metrics-collector/internal/model"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.statusCode = statusCode
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// WithAudit middleware sends audit events after successful metrics processing.
func WithAudit(logger *zap.Logger, publisher *audit.Publisher) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read request body for audit",
					zap.Error(err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			start := time.Now()
			arw := &auditResponseWriter{ResponseWriter: w}

			next.ServeHTTP(arw, r)

			if arw.statusCode < 200 || arw.statusCode >= 300 {
				return
			}

			metrics := extractMetrics(r, body)
			if len(metrics) == 0 {
				return
			}

			publisher.Notify(audit.Event{
				Timestamp: start.Unix(),
				Metrics:      metrics,
				IPAddress: r.RemoteAddr,
			})
		})
	}
}

func extractMetrics(r *http.Request, body []byte) []string {
	if name := chi.URLParam(r, "name"); name != "" {
		return []string{name}
	}

	if len(body) == 0 {
		return nil
	}

	var batch []models.Metrics
	if json.NewDecoder(bytes.NewReader(body)).Decode(&batch) == nil {
		names := make([]string, len(batch))
		for i, m := range batch {
			names[i] = m.ID
		}
		return names
	}

	var single models.Metrics
	if json.NewDecoder(bytes.NewReader(body)).Decode(&single) == nil {
		return []string{single.ID}
	}

	return nil
}
