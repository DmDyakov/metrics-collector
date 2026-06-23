package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"metrics-collector/internal/handler/mocks"
	models "metrics-collector/internal/model"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func ExampleMetricsHandler_UpdateMetricByJSON() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	value := 42.5
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
		Value: &value,
	}

	mockService.EXPECT().
		UpdateMetric(gomock.Any(), metric).
		Return(&metric, nil)

	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	body, _ := json.Marshal(metric)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateMetricByJSON(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleMetricsHandler_UpdateMetricByJSON_invalidJSON() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	body := []byte(`{not valid json`)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateMetricByJSON(rec, req)

	fmt.Println(rec.Code)
	// Output: 400
}

func ExampleMetricsHandler_ListMetrics() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	value := 42.5
	metrics := []models.Metrics{
		{ID: "cpu", MType: "gauge", Value: &value},
	}

	mockService.EXPECT().GetAllMetrics().Return(metrics, nil)

	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ListMetrics(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleMetricsHandler_GetMetricValue() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	value := 42.5
	metric := &models.Metrics{ID: "cpu", MType: "gauge", Value: &value}

	mockService.EXPECT().GetMetric(gomock.Any()).Return(metric, nil)

	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "cpu")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	h.GetMetricValue(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleMetricsHandler_GetMetric() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	value := 42.5
	metric := &models.Metrics{ID: "cpu", MType: "gauge", Value: &value}

	mockService.EXPECT().GetMetric(gomock.Any()).Return(metric, nil)

	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	body := `{"id":"cpu","type":"gauge"}`
	req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.GetMetric(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleMetricsHandler_UpdateMetricByURL() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	value := 42.5
	metric := &models.Metrics{ID: "cpu", MType: "gauge", Value: &value}

	mockService.EXPECT().UpdateMetric(gomock.Any(), gomock.Any()).Return(metric, nil)

	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/cpu/42.5", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "cpu")
	rctx.URLParams.Add("value", "42.5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	h.UpdateMetricByURL(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleMetricsHandler_UpdateMetricsBatch() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := mocks.NewMockMetricsService(ctrl)
	count := 3

	mockService.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).Return(&count, nil)

	h, _ := NewMetricsHandler(mockService, zap.NewNop())

	body := `[{"id":"cpu","type":"gauge","value":1.1},{"id":"mem","type":"gauge","value":2.2},{"id":"hits","type":"counter","delta":10}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateMetricsBatch(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}
