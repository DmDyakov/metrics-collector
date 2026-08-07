package client

import (
	"metrics-collector/internal/domain/metrics"
	"metrics-collector/internal/dto"
)

func (c *Client) toDto(batch map[string]float64) []dto.Metric {
	dtoBatch := make([]dto.Metric, 0, len(batch))

	for k, v := range batch {
		if k == metrics.PollCount {
			delta := int64(v)
			dtoBatch = append(dtoBatch, dto.Metric{
				ID:    k,
				Type:  metrics.Counter,
				Delta: &delta,
			})
		} else {
			value := v
			dtoBatch = append(dtoBatch, dto.Metric{
				ID:    k,
				Type:  metrics.Gauge,
				Value: &value,
			})
		}
	}

	return dtoBatch
}
