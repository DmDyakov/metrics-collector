package handler

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"metrics-collector/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	service *service.MetricsService
}

func NewHandler(service *service.MetricsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RootHandle(res http.ResponseWriter, req *http.Request) {
	allMetrics, err := h.service.GetAllMetrics()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	allMetricsHTMLTemplate, err := template.ParseFiles("internal/templates/metrics.html")
	if err != nil {
		http.Error(res, "template not found", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := allMetricsHTMLTemplate.Execute(&buf, allMetrics); err != nil {
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	buf.WriteTo(res)
}

func (h *Handler) ValueHandle(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	metricName := chi.URLParam(req, "name")

	if metricType == "" {
		http.Error(res, "metric type is required", http.StatusNotFound)
		return
	}

	if metricName == "" {
		http.Error(res, "metric name is required", http.StatusNotFound)
		return
	}

	value, err := h.service.GetMetricValue(metricType, metricName)
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

	res.WriteHeader(http.StatusOK)
	io.WriteString(res, value)

}

func (h *Handler) UpdateHandle(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	metricName := chi.URLParam(req, "name")
	metricValue := chi.URLParam(req, "value")

	if metricType == "" {
		http.Error(res, "metric type is required", http.StatusNotFound)
		return
	}

	if metricName == "" {
		http.Error(res, "metric name is required", http.StatusNotFound)
		return
	}

	if metricValue == "" {
		http.Error(res, "metric value is required", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMetric(metricType, metricName, metricValue); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handler) NewMetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/", func(r chi.Router) {
		r.Get("/", h.RootHandle)                     // GET /
		r.Get("/value/{type}/{name}", h.ValueHandle) // GET /value/gauge/RandomValue
	})
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandle) // POST /value/gauge/RandomValue/123.456
	return r
}
