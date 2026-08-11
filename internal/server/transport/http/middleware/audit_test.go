package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics-collector/internal/domain/audit"
	"metrics-collector/internal/server/transport/http/middleware/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestWithAudit(t *testing.T) {
	t.Run("publishes event on successful metrics update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var receivedEvent audit.Event
		mockPub := mocks.NewMockAuditPublisher(ctrl)
		mockPub.EXPECT().Publish(gomock.Any()).Do(func(e audit.Event) {
			receivedEvent = e
		})

		handler := WithAudit(zap.NewNop(), mockPub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		metric := map[string]interface{}{
			"id":    "cpu",
			"type":  "gauge",
			"value": 42.5,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "cpu", receivedEvent.Metrics[0])
		assert.Equal(t, "192.168.1.1:12345", receivedEvent.IPAddress)
	})

	t.Run("publishes event for batch update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var receivedEvent audit.Event
		mockPub := mocks.NewMockAuditPublisher(ctrl)
		mockPub.EXPECT().Publish(gomock.Any()).Do(func(e audit.Event) {
			receivedEvent = e
		})

		handler := WithAudit(zap.NewNop(), mockPub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		batch := []map[string]interface{}{
			{"id": "cpu", "type": "gauge", "value": 42.5},
			{"id": "mem", "type": "gauge", "value": 70.0},
		}
		body, _ := json.Marshal(batch)

		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.RemoteAddr = "10.0.0.1:54321"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, receivedEvent.Metrics, 2)
		assert.Equal(t, "cpu", receivedEvent.Metrics[0])
		assert.Equal(t, "mem", receivedEvent.Metrics[1])
	})

	t.Run("publishes event from chi URL param", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var receivedEvent audit.Event
		mockPub := mocks.NewMockAuditPublisher(ctrl)
		mockPub.EXPECT().Publish(gomock.Any()).Do(func(e audit.Event) {
			receivedEvent = e
		})

		handler := WithAudit(zap.NewNop(), mockPub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu", nil)
		req.RemoteAddr = "172.16.0.1:9999"

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "cpu")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "cpu", receivedEvent.Metrics[0])
	})

	t.Run("does not publish on error status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPub := mocks.NewMockAuditPublisher(ctrl)

		handler := WithAudit(zap.NewNop(), mockPub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte(`{"id":"cpu"}`)))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestExtractMetrics(t *testing.T) {
	t.Run("extracts from chi URL param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("name", "cpu")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		metrics := extractMetrics(req, nil)
		assert.Equal(t, []string{"cpu"}, metrics)
	})

	t.Run("extracts from batch body", func(t *testing.T) {
		batch := []map[string]interface{}{
			{"id": "cpu", "type": "gauge", "value": 42.5},
			{"id": "mem", "type": "gauge", "value": 70.0},
		}
		body, _ := json.Marshal(batch)

		req := httptest.NewRequest(http.MethodPost, "/updates", nil)
		metrics := extractMetrics(req, body)
		assert.Equal(t, []string{"cpu", "mem"}, metrics)
	})

	t.Run("extracts from single metric body", func(t *testing.T) {
		metric := map[string]interface{}{
			"id":    "cpu",
			"type":  "gauge",
			"value": 42.5,
		}
		body, _ := json.Marshal(metric)

		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		metrics := extractMetrics(req, body)
		assert.Equal(t, []string{"cpu"}, metrics)
	})

	t.Run("returns nil for empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		metrics := extractMetrics(req, nil)
		assert.Nil(t, metrics)
	})

	t.Run("returns nil for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		metrics := extractMetrics(req, []byte("not json"))
		assert.Nil(t, metrics)
	})
}
