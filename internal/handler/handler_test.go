package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics-collector/internal/handler"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"

	"go.uber.org/zap"
)

func TestUpdateMetricV2(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)

	h, err := handler.NewHandler(svc, logger)
	if err != nil {
		t.Logf("could not load template: %v", err)
	}

	router := h.NewMetricsRouter()

	tests := []struct {
		name       string
		metric     models.Metrics
		wantCode   int
		wantErrMsg string
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
			wantCode:   http.StatusBadRequest,
			wantErrMsg: "unknown metric type: unknown\n",
		},
		{
			name: "negative - missing delta for counter",
			metric: models.Metrics{
				ID:    "Test",
				MType: models.Counter,
			},
			wantCode:   http.StatusBadRequest,
			wantErrMsg: "metric delta is required for counter\n",
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

			if tt.wantErrMsg != "" && w.Body.String() != tt.wantErrMsg {
				t.Errorf("response body = %q, want %q", w.Body.String(), tt.wantErrMsg)
			}
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	testTable := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "positive test #1 - update counter metric",
			url:  "/update/counter/testPollCount/1",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "positive test #2 - update gauge metric",
			url:  "/update/gauge/testRandomValue/123.456",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "invalid URL format",
			url:  "/update/invalid",
			want: want{
				code:        http.StatusNotFound,
				response:    "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "unknown metric type",
			url:  "/update/unknown/test/123",
			want: want{
				code:        http.StatusBadRequest,
				response:    "unknown metric type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	logger := zap.NewNop().Sugar()
	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	h, err := handler.NewHandler(svc, logger)
	if err != nil {
		t.Logf("could not load template: %v", err)
	}

	router := h.NewMetricsRouter()

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.want.code {
				t.Errorf("UpdateMetric() status = %v, want %v", w.Code, tt.want)
			}

			if w.Body.String() != tt.want.response {
				t.Errorf("response body = %q, want %q", w.Body.String(), tt.want.response)
			}

			if tt.want.contentType != "" {
				if ct := w.Header().Get("Content-Type"); ct != tt.want.contentType {
					t.Errorf("Content-Type = %q, want %q", ct, tt.want.contentType)
				}
			}
		})
	}
}
