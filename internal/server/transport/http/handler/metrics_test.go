package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metrics-collector/internal/server/errs"
	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/transport/http/handler/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestMetricsHandler_HandleError(t *testing.T) {
	h := &MetricsHandler{logger: zap.NewNop()}

	t.Run("bad request: invalid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, errs.ErrInvalidRequest)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request: unknown metric type", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, errs.ErrUnknownMetricType)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request: invalid counter value", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, errs.ErrInvalidCounterValue)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad request: invalid gauge value", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, errs.ErrInvalidGaugeValue)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, &errs.MetricNotFoundError{Type: "gauge", Name: "cpu"})
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, errs.ErrInvalidResponse)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("unknown error defaults to 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.handleError(w, errors.New("something unexpected"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMetricsHandler_GetMetricValue(t *testing.T) {
	t.Run("gauge returns formatted value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockMetricsService(ctrl)
		h, _ := NewMetricsHandler(mockService, zap.NewNop())

		value := 42.5
		metric := &models.Metrics{ID: "cpu", MType: models.Gauge, Value: &value}
		mockService.EXPECT().GetMetric(gomock.Any()).Return(metric, nil)

		req := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu", nil)
		w := httptest.NewRecorder()

		h.GetMetricValue(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "42.5", w.Body.String())
	})

	t.Run("counter returns formatted value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockMetricsService(ctrl)
		h, _ := NewMetricsHandler(mockService, zap.NewNop())

		delta := int64(10)
		metric := &models.Metrics{ID: "hits", MType: models.Counter, Delta: &delta}
		mockService.EXPECT().GetMetric(gomock.Any()).Return(metric, nil)

		req := httptest.NewRequest(http.MethodGet, "/value/counter/hits", nil)
		w := httptest.NewRecorder()

		h.GetMetricValue(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "10", w.Body.String())
	})
}

func TestMetricsHandler_UpdateMetricByURL(t *testing.T) {
	t.Run("gauge from url params", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockMetricsService(ctrl)
		h, _ := NewMetricsHandler(mockService, zap.NewNop())

		value := 42.5
		expected := &models.Metrics{ID: "cpu", MType: "gauge", Value: &value}
		mockService.EXPECT().UpdateMetric(gomock.Any(), gomock.Any()).Return(expected, nil)

		req := httptest.NewRequest(http.MethodPost, "/update/gauge/cpu/42.5", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("type", "gauge")
		rctx.URLParams.Add("name", "cpu")
		rctx.URLParams.Add("value", "42.5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.UpdateMetricByURL(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid counter value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockMetricsService(ctrl)
		h, _ := NewMetricsHandler(mockService, zap.NewNop())

		req := httptest.NewRequest(http.MethodPost, "/update/counter/hits/abc", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("type", "counter")
		rctx.URLParams.Add("name", "hits")
		rctx.URLParams.Add("value", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.UpdateMetricByURL(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestMetricsHandler_GetMetric(t *testing.T) {
	t.Run("returns metric by JSON", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockMetricsService(ctrl)
		h, _ := NewMetricsHandler(mockService, zap.NewNop())

		value := 42.5
		metric := &models.Metrics{ID: "cpu", MType: "gauge", Value: &value}
		mockService.EXPECT().GetMetric(gomock.Any()).Return(metric, nil)

		body := `{"id":"cpu","type":"gauge"}`
		req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.GetMetric(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
