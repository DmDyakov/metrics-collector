package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"metrics-collector/internal/domain/metrics"
	"metrics-collector/internal/server/errs"
	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_metrics_service.go -package=mocks . MetricsService
type MetricsService interface {
	UpdateMetric(ctx context.Context, m models.Metrics) (*models.Metrics, error)
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) (*int, error)
	GetMetric(m models.Metrics) (*models.Metrics, error)
	GetAllMetrics() ([]models.Metrics, error)
}

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
	if _, err := buf.WriteTo(w); err != nil {
		h.logger.Error("failed to write HTML response", zap.Error(err))
	}
}

// GetMetricValue возвращает значение метрики.
func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := metrics.Type(chi.URLParam(r, "type"))
	metricName := chi.URLParam(r, "name")

	m := models.Metrics{ID: metricName, MType: metricType}

	metric, err := h.service.GetMetric(m)
	if err != nil {
		h.handleError(w, err)
		return
	}

	var value string
	switch metric.MType {
	case metrics.Counter:
		value = strconv.FormatInt(*metric.Delta, 10)
	case metrics.Gauge:
		value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, value); err != nil {
		h.logger.Error("failed to write metric value", zap.Error(err))
	}
}

// GetMetric возвращает метрику.
func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.handleError(w, fmt.Errorf("%w: invalid JSON body", errs.ErrInvalidRequest))
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
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricByURL обновляет метрику через URL-параметры.
func (h *MetricsHandler) UpdateMetricByURL(w http.ResponseWriter, r *http.Request) {
	metricType := metrics.Type(chi.URLParam(r, "type"))
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	m := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case metrics.Counter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			h.handleError(w, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, errs.ErrInvalidCounterValue))
			return
		}
		m.Delta = &delta

	case metrics.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			h.handleError(w, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, errs.ErrInvalidGaugeValue))
			return
		}
		m.Value = &value

	default:
		h.handleError(w, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, errs.ErrUnknownMetricType))
		return
	}

	updated, err := h.service.UpdateMetric(r.Context(), m)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(updated); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricByJSON обновляет метрику через JSON-запрос.
func (h *MetricsHandler) UpdateMetricByJSON(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.handleError(w, fmt.Errorf("%w: invalid JSON body", errs.ErrInvalidRequest))
		return
	}

	updated, err := h.service.UpdateMetric(r.Context(), m)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(updated); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricsBatch пакетно обновляет метрики.
func (h *MetricsHandler) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		h.handleError(w, fmt.Errorf("%w: invalid JSON body", errs.ErrInvalidRequest))
		return
	}

	if len(metrics) == 0 {
		h.handleError(w, fmt.Errorf("%w: empty metrics array", errs.ErrInvalidRequest))
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

	case errors.Is(err, errs.ErrInvalidRequest),
		errors.Is(err, errs.ErrUnknownMetricType),
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
