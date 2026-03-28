package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RootHandle(w http.ResponseWriter, r *http.Request) {
	allMetrics, err := h.service.GetAllMetrics()
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var buf bytes.Buffer
	if err := h.allMetricsHTMLTemplate.Execute(&buf, allMetrics); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, *value)

}

func (h *Handler) UpdateHandle(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if err := h.service.UpdateMetric(metricType, metricName, metricValue); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
