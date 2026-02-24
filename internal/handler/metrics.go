package handler

import (
	"net/http"
	"strings"

	"metrics-collector/internal/service"
)

type Handler struct {
	service *service.MetricsService
}

func NewHandler(service *service.MetricsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateMetrics(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/update/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		http.Error(res, "invalid URL", http.StatusNotFound)
		return
	}

	metricType, metricName, metricValue := parts[0], parts[1], parts[2]

	if metricName == "" {
		http.Error(res, "metric name is required", http.StatusNotFound)
		return
	}

	if metricValue == "" {
		http.Error(res, "metric value is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(metricType, metricName, metricValue); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}
