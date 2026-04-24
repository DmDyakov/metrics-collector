package agent

import (
	"encoding/json"
	models "metrics-collector/internal/model"

	"go.uber.org/zap"
)

type metricRequest struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"` // для Counter
	Value *float64 `json:"value,omitempty"` // для Gauge
}

func buildBatch(m rawMetrics) []metricRequest {
	batch := make([]metricRequest, 0, len(m))
	for k, v := range m {
		if k == "PollCount" {
			delta := int64(v)
			batch = append(batch, metricRequest{
				ID:    k,
				Type:  models.Counter,
				Delta: &delta,
			})
		} else {
			value := v
			batch = append(batch, metricRequest{
				ID:    k,
				Type:  models.Gauge,
				Value: &value,
			})
		}
	}
	return batch
}

func (a *Agent) compress(batch []metricRequest) ([]byte, error) {
	jsonPayload, err := json.Marshal(batch)
	if err != nil {
		a.logger.Error("error JSON marshaling", zap.Error(err))
		return nil, err
	}
	compressedJson, err := a.gzip.Compress(jsonPayload)
	if err != nil {
		a.logger.Error("error JSON compressing", zap.Error(err))
		return nil, err
	}

	return compressedJson, nil
}
