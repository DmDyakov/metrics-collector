package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

type contextKey string

const (
	startTimeKey contextKey = "startTime"
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.responseData.status == 0 {
		r.responseData.status = http.StatusOK
	}

	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if r.responseData.status == 0 {
		r.responseData.status = statusCode
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// WithLogging middleware logs all incoming HTTP requests and responses.
func WithLogging(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			responseData := &responseData{
				status: 0,
				size:   0,
			}

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			next.ServeHTTP(lw, r)

			if responseData.status == 0 {
				responseData.status = http.StatusOK
			}

			fields := []zap.Field{
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
			}

			if startTime, ok := r.Context().Value(startTimeKey).(time.Time); ok {
				duration := time.Since(startTime)
				fields = append(fields, zap.Duration("duration", duration))
			}

			logger.Info("Response", fields...)
		})
	}
}
