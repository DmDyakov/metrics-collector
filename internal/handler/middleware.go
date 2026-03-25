package handler

import (
	"net/http"

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

func WithLogging(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(lw, r)

		if responseData.status == 0 {
			responseData.status = http.StatusOK
		}

		logger.Infoln(
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}
