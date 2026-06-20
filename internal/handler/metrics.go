package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"metrics-collector/internal/errs"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/templates"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_metrics_service.go -package=mocks . MetricsService
type MetricsService interface {
	UpdateMetricByArgs(ctx context.Context, metricType, metricName, metricValue string) (*models.Metrics, error)
	UpdateMetricByJSON(ctx context.Context, metric models.Metrics) (*models.Metrics, error)
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) (*int, error)
	GetMetricValueByURL(metricType, metricName string) (*string, error)
	GetMetric(m models.Metrics) (*models.Metrics, error)
	GetAllMetrics() ([]models.Metrics, error)
}

// MetricsHandler обрабатывает HTTP-запросы для сервиса сбора метрик.
type MetricsHandler struct {
	service                MetricsService
	logger                 *zap.Logger
	allMetricsHTMLTemplate *template.Template
}

// NewMetricsHandler создаёт новый MetricsHandler.
func NewMetricsHandler(
	service MetricsService,
	logger *zap.Logger,
) (*MetricsHandler, error) {
	tmpl, err := template.ParseFS(templates.FS, "metrics.html")
	if err != nil {
		return nil, err
	}

	return &MetricsHandler{
		service:                service,
		logger:                 logger,
		allMetricsHTMLTemplate: tmpl,
	}, nil
}

// ListMetrics возвращает HTML-страницу со списком всех метрик.
func (h *MetricsHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics, err := h.service.GetAllMetrics()
	if err != nil {
		h.handleError(w, err)
		return
	}

	var buf bytes.Buffer
	if err := h.allMetricsHTMLTemplate.Execute(&buf, allMetrics); err != nil {
		h.logger.Error("failed to execute metrics HTML template", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

// GetMetricValue возвращает значение метрики.
func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	value, err := h.service.GetMetricValueByURL(metricType, metricName)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, *value)

}

// GetMetric возвращает метрику.
func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(m)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(metric); err != nil {
		http.Error(w, "invalid JSON body", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricByURL обновляет метрику через URL-параметры.
func (h *MetricsHandler) UpdateMetricByURL(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	updated, err := h.service.UpdateMetricByArgs(r.Context(), metricType, metricName, metricValue)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(updated); err != nil {
		http.Error(w, "invalid JSON body", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricByJSON обновляет метрику через JSON-запрос.
func (h *MetricsHandler) UpdateMetricByJSON(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	updatedMetric, err := h.service.UpdateMetricByJSON(r.Context(), m)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(updatedMetric); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricsBatch пакетно обновляет метрики.
func (h *MetricsHandler) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		http.Error(w, "empty metrics array", http.StatusBadRequest)
		return
	}

	count, err := h.service.UpdateMetrics(r.Context(), metrics)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(map[string]int{
		"updated": *count,
		"total":   len(metrics),
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleError обрабатывает ошибки сервиса и возвращает соответствующий HTTP-статус.
func (h *MetricsHandler) handleError(w http.ResponseWriter, err error) {
	var errMetricNotFound *errs.MetricNotFoundError

	switch {
	case errors.Is(err, errs.ErrInvalidResponse):
		h.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	case errors.As(err, &errMetricNotFound):
		http.Error(w, errMetricNotFound.Error(), http.StatusNotFound)

	case errors.Is(err, errs.ErrUnknownMetricType),
		errors.Is(err, errs.ErrMetricTypeMismatch),
		errors.Is(err, errs.ErrInvalidCounterValue),
		errors.Is(err, errs.ErrInvalidGaugeValue),
		errors.Is(err, errs.ErrMetricDeltaForCountRequired):
		h.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

	default:
		h.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
