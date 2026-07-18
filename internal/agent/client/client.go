package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"metrics-collector/internal/agent/compress"

	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
	gzip       *compress.Gzip
	logger     *zap.Logger
}

func New(baseURL, secretKey string, logger *zap.Logger, gzip *compress.Gzip) *Client {
	return &Client{
		baseURL:   baseURL,
		secretKey: secretKey,
		gzip:      gzip,
		logger:    logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendMetrics(ctx context.Context, batch map[string]float64) error {
	if len(batch) <= 0 {
		c.logger.Warn("No metrics for send")
		return nil
	}

	c.logger.Info("Try to send metrics")
	start := time.Now()

	url := fmt.Sprintf("http://%s/updates", c.baseURL)
	method := http.MethodPost
	reqBody, err := c.compress(c.toDto(batch))
	if err != nil {
		c.logger.Warn("Failed to compress metrics, skipping sending", zap.Error(err))
		return nil
	}

	doRequest := func() (*http.Response, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
		if err != nil {
			c.logger.Error("failed to build request",
				zap.String("uri", url),
				zap.String("method", method),
				zap.Error(err),
			)
			return nil, err
		}

		if c.secretKey != "" {
			req.Header.Set("HashSHA256", hex.EncodeToString(c.createSignature(reqBody)))
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp != nil {
			defer resp.Body.Close()

			if c.secretKey != "" {
				err = c.checkResponseSignature(resp)
			}
		}

		return resp, err
	}

	if err := c.withRetry(ctx, doRequest); err != nil {
		c.logger.Warn("Failed to send request after retries",
			zap.String("uri", url),
			zap.String("method", method),
			zap.Error(err),
		)
		return err
	} else {
		duration := time.Since(start)
		c.logger.Info("request sent",
			zap.String("uri", url),
			zap.String("method", "POST"),
			zap.Duration("duration", duration),
		)
	}

	return nil
}
