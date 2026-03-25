package handler

import (
	"encoding/json"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/service"
	"net/http"
)

func (h *Handler) ValueHandleV2(res http.ResponseWriter, req *http.Request) {
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

	metric, err := h.service.GetMetricValueV2(m)
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

func (h *Handler) UpdateHandleV2(res http.ResponseWriter, req *http.Request) {
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

	updatedMetric, err := h.service.UpdateMetricV2(m)
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
