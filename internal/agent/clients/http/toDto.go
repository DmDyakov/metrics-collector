package client

import (
	"metrics-collector/internal/dto"
	models "metrics-collector/internal/server/model"
)

func (c *Client) toDto(metrics map[string]float64) []dto.Metric {
	dtoBatch := make([]dto.Metric, 0, len(metrics))

	for k, v := range metrics {
		if k == models.PollCount {
			delta := int64(v)
			dtoBatch = append(dtoBatch, dto.Metric{
				ID:    k,
				Type:  models.Counter,
				Delta: &delta,
			})
		} else {
			value := v
			dtoBatch = append(dtoBatch, dto.Metric{
				ID:    k,
				Type:  models.Gauge,
				Value: &value,
			})
		}
	}

	return dtoBatch
}
