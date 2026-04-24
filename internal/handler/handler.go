package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"time"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"
	"metrics-collector/internal/errs"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/templates"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricsService interface {
	UpdateMetricByArgs(ctx context.Context, metricType, metricName, metricValue string) (*models.Metrics, error)
	UpdateMetricByJSON(ctx context.Context, metric models.Metrics) (*models.Metrics, error)
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) (*int, error)
	GetMetricValue(metricType, metricName string) (*string, error)
	GetMetric(m models.Metrics) (*models.Metrics, error)
	GetAllMetrics() ([]models.Metrics, error)
	Ping(ctx context.Context) error
}

type Handler struct {
	service                MetricsService
	logger                 *zap.Logger
	gzip                   *compress.Gzip
	allMetricsHTMLTemplate *template.Template
	secretKey              string
}

func NewHandler(
	service MetricsService,
	logger *zap.Logger,
	gzip *compress.Gzip,
	cfg *config.ServerConfig,
) (*Handler, error) {
	tmpl, err := template.ParseFS(templates.FS, "metrics.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		service:                service,
		logger:                 logger,
		gzip:                   gzip,
		allMetricsHTMLTemplate: tmpl,
		secretKey:              cfg.SecretKey,
	}, nil
}

func (h *Handler) NewMetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(h.WithLogging)
	r.Use(h.WithSignature)
	r.Use(h.WithCompressing)

	r.Get("/", h.RootHandle)
	r.Get("/ping", h.PingHandle)
	r.Get("/value/{type}/{name}", h.ValueHandle)
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandle)
	r.Post("/value", h.ValueByJSONHandle)
	r.Post("/update", h.UpdateByJSONHandle)
	r.Post("/updates", h.UpdatesHandle)

	return r
}

func (h *Handler) PingHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.service.Ping(ctx)
	if err != nil {
		h.logger.Error("Database ping failed", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *Handler) RootHandle(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ValueHandle(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	value, err := h.service.GetMetricValue(metricType, metricName)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, *value)

}

func (h *Handler) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updated, err := h.service.UpdateMetricByArgs(ctx, metricType, metricName, metricValue)
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

func (h *Handler) ValueByJSONHandle(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateByJSONHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var m models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	updatedMetric, err := h.service.UpdateMetricByJSON(ctx, m)
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

func (h *Handler) UpdatesHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	count, err := h.service.UpdateMetrics(ctx, metrics)
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

func (h *Handler) handleError(w http.ResponseWriter, err error) {
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
