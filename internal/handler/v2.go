package handler

import (
	"encoding/json"
	models "metrics-collector/internal/model"
	"net/http"
)

func (h *Handler) ValueHandleV2(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetricValueV2(m)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(metric); err != nil {
		http.Error(w, "invalid JSON body", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateHandleV2(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	updatedMetric, err := h.service.UpdateMetricV2(m)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(updatedMetric); err != nil {
		http.Error(w, "invalid JSON body", http.StatusInternalServerError)
		return
	}
}
