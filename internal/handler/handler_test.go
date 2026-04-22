package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/handler/mocks"
	models "metrics-collector/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

//go:generate mockgen -destination=mocks/mock_service.go -package=mocks . MetricsService
func TestHandler_UpdateHandleJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name            string
		body            []byte
		setupMock       func(*mocks.MockMetricsService)
		wantCode        int
		wantContentType string
	}{
		{
			name: "positive: update counter metric",
			body: []byte(`{
				"id": "TestCount",
				"type": "counter",
				"value": 10
			}`),
			setupMock: func(m *mocks.MockMetricsService) {
				m.EXPECT().
					UpdateMetricByJSON(gomock.Any(), gomock.Any()).
					Return(&models.Metrics{}, nil).
					Times(1)
			},
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name: "positive: update gauge metric",
			body: []byte(`{
				"id": "TestGauge",
				"type": "gauge",
				"value": 123.456
			}`),
			setupMock: func(m *mocks.MockMetricsService) {
				m.EXPECT().
					UpdateMetricByJSON(gomock.Any(), gomock.Any()).
					Return(&models.Metrics{}, nil).
					Times(1)
			},
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name: "negative: empty body",
			body: []byte(""),
			setupMock: func(m *mocks.MockMetricsService) {
				m.EXPECT().
					UpdateMetricByJSON(gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantCode:        http.StatusBadRequest,
			wantContentType: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gzip := compress.NewGzip()
			mockSvc := mocks.NewMockMetricsService(ctrl)

			h, err := NewHandler(mockSvc, logger, gzip)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(tt.body))

			tt.setupMock(mockSvc)

			w := httptest.NewRecorder()
			h.UpdateByJSONHandle(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
		})
	}
}
