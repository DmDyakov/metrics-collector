package handler

import (
	"bytes"
	"io"
	"metrics-collector/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

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
