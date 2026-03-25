package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	models "metrics-collector/internal/model"
	"metrics-collector/internal/service"
	"metrics-collector/internal/templates"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	service                *service.MetricsService
	logger                 *zap.SugaredLogger
	allMetricsHTMLTemplate *template.Template
}

func NewHandler(service *service.MetricsService, logger *zap.SugaredLogger) (*Handler, error) {
	tmpl, err := template.ParseFS(templates.FS, "metrics.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		service:                service,
		logger:                 logger,
		allMetricsHTMLTemplate: tmpl,
	}, nil
}

func (h *Handler) RootHandle(res http.ResponseWriter, req *http.Request) {
	allMetrics, err := h.service.GetAllMetrics()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := h.allMetricsHTMLTemplate.Execute(&buf, allMetrics); err != nil {
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	buf.WriteTo(res)
}

func (h *Handler) ValueHandle(res http.ResponseWriter, req *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(res, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}
	if err := m.ValidateBase(); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetricValue(m)
	if err != nil {
		switch {
		case err == service.ErrMetricNotFound:
			http.Error(res, err.Error(), http.StatusNotFound)
		case err == service.ErrUnknownMetricType:
			http.Error(res, err.Error(), http.StatusBadRequest)
		default:
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
		return
	}

	res.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(res)
	if err := enc.Encode(metric); err != nil {
		http.Error(res, "error encoding response", http.StatusBadRequest)
		return
	}
}

func (h *Handler) UpdateHandle(res http.ResponseWriter, req *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(res, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	if err := m.ValidateForUpdate(); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	updatedMetric, err := h.service.UpdateMetric(m)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(res)
	if err := enc.Encode(updatedMetric); err != nil {
		http.Error(res, "error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) NewMetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return WithLogging(next, h.logger)
	})

	r.Get("/", h.RootHandle)
	r.Post("/value", h.ValueHandle)
	r.Post("/update", h.UpdateHandle)
	return r
}
