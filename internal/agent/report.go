package agent

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (a *Agent) reportBatch(batch []batchMetric) {
	if len(batch) <= 0 {
		a.logger.Warn("No metrics for reporting")
		return
	}

	a.logger.Info("Metrics reporting...")
	start := time.Now()

	url := fmt.Sprintf("http://%s/updates", a.cfg.ServerBaseURL)
	method := "POST"
	reqBody, err := a.compress(batch)
	if err != nil {
		a.logger.Error("error compress batch", zap.Error(err))
		return
	}

	doRequest := func() (*http.Response, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
		if err != nil {
			a.logger.Error("failed to build request",
				zap.String("uri", url),
				zap.String("method", method),
				zap.Error(err),
			)
			return nil, err
		}

		if a.cfg.SecretKey != "" {
			req.Header.Set("HashSHA256", hex.EncodeToString(a.createSignature(reqBody)))
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := a.client.Do(req)
		if resp != nil {
			defer resp.Body.Close()

			if a.cfg.SecretKey != "" {
				err = a.checkResponseSignature(resp)
			}
		}

		return resp, err
	}

	if err := a.withRetry(doRequest); err != nil {
		a.logger.Error("failed to send request",
			zap.String("uri", url),
			zap.String("method", method),
			zap.Error(err),
		)
	} else {
		duration := time.Since(start)
		a.logger.Info("request sent",
			zap.String("uri", url),
			zap.String("method", "POST"),
			zap.Duration("duration", duration),
		)
	}
}
