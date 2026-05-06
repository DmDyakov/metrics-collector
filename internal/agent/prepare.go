package agent

import (
	"encoding/json"
	models "metrics-collector/internal/model"

	"go.uber.org/zap"
)

func (a *Agent) buildReportingBatch() []batchMetric {
	a.store.mu.RLock()

	batch := make([]batchMetric, 0, len(a.store.metrics))
	for k, v := range a.store.metrics {
		if k == PollCount {
			delta := int64(v)
			batch = append(batch, batchMetric{
				ID:    k,
				Type:  models.Counter,
				Delta: &delta,
			})
		} else {
			value := v
			batch = append(batch, batchMetric{
				ID:    k,
				Type:  models.Gauge,
				Value: &value,
			})
		}
	}

	a.store.mu.RUnlock()

	return batch
}

func (a *Agent) compress(batch []batchMetric) ([]byte, error) {
	jsonPayload, err := json.Marshal(batch)
	if err != nil {
		a.logger.Error("error JSON marshaling", zap.Error(err))
		return nil, err
	}
	compressedJSON, err := a.gzip.Compress(jsonPayload)
	if err != nil {
		a.logger.Error("error JSON compressing", zap.Error(err))
		return nil, err
	}

	return compressedJSON, nil
}
