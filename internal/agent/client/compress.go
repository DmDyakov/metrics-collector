package client

import (
	"encoding/json"
	"metrics-collector/internal/dto"

	"go.uber.org/zap"
)

func (c *Client) compress(batch []dto.Metric) ([]byte, error) {
	jsonPayload, err := json.Marshal(batch)
	if err != nil {
		c.logger.Error("error JSON marshaling", zap.Error(err))
		return nil, err
	}
	compressedJSON, err := c.gzip.Compress(jsonPayload)
	if err != nil {
		c.logger.Error("error JSON compressing", zap.Error(err))
		return nil, err
	}

	return compressedJSON, nil
}
