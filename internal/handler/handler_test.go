package handler_test

import (
	"metrics-collector/internal/handler"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

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
	h := handler.NewHandler(svc, logger)

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
