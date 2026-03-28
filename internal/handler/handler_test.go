package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/handler"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"

	"go.uber.org/zap"
)

func TestHandler_UpdateMetricV2(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	gzip := compress.NewGzip()
	h, err := handler.NewHandler(svc, logger, gzip)
	if err != nil {
		t.Logf("could not load template: %v", err)
	}

	router := h.NewMetricsRouter()

	tests := []struct {
		name     string
		metric   models.Metrics
		wantCode int
	}{
		{
			name: "positive - update counter metric",
			metric: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: func() *int64 { v := int64(1); return &v }(),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "positive - update gauge metric",
			metric: models.Metrics{
				ID:    "RandomValue",
				MType: models.Gauge,
				Value: func() *float64 { v := 123.456; return &v }(),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "negative - invalid metric type",
			metric: models.Metrics{
				ID:    "Test",
				MType: "unknown",
				Delta: func() *int64 { v := int64(1); return &v }(),
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "negative - missing delta for counter",
			metric: models.Metrics{
				ID:    "Test",
				MType: models.Counter,
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.metric)
			if err != nil {
				t.Fatalf("failed to marshal metric: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("status code = %v, want %v", w.Code, tt.wantCode)
			}
		})
	}
}

func TestHandler_UpdateMetric(t *testing.T) {

	testTable := []struct {
		name     string
		url      string
		wantCode int
	}{
		{
			name:     "positive test #1 - update counter metric",
			url:      "/update/counter/testPollCount/1",
			wantCode: http.StatusOK,
		},
		{
			name:     "positive test #2 - update gauge metric",
			url:      "/update/gauge/testRandomValue/123.456",
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid URL format",
			url:      "/update/invalid",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "unknown metric type",
			url:      "/update/unknown/test/123",
			wantCode: http.StatusBadRequest,
		},
	}

	logger := zap.NewNop().Sugar()
	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	gzip := compress.NewGzip()
	h, err := handler.NewHandler(svc, logger, gzip)
	if err != nil {
		t.Logf("could not load template: %v", err)
	}

	router := h.NewMetricsRouter()

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("UpdateMetric() status = %v, want %v", w.Code, tt.wantCode)
			}
		})
	}
}
