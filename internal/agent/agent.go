package agent

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"

	"go.uber.org/zap"
)

type rawMetrics map[string]float64

type Agent struct {
	cfg     *config.AgentConfig
	logger  *zap.Logger
	gzip    *compress.Gzip
	client  *http.Client
	metrics map[string]float64
}

func NewAgent(cfg *config.AgentConfig, logger *zap.Logger, gzip *compress.Gzip) *Agent {
	return &Agent{
		cfg:    cfg,
		logger: logger,
		gzip:   gzip,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		metrics: make(rawMetrics),
	}
}

func (a *Agent) Run() {
	reportMultiplier := int64(a.cfg.ReportInterval / a.cfg.PollInterval)
	var pollCount int64 = 0

	pollTicker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()

	for range pollTicker.C {
		pollCount++

		a.poll(pollCount)

		if pollCount%reportMultiplier == 0 {
			a.send()
		}
	}

}

func (a *Agent) poll(count int64) {
	a.logger.Info("Опрос метрик",
		zap.Int64("iteration", count),
	)

	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	a.metrics["PollCount"] = float64(count)
	a.metrics["RandomValue"] = rand.Float64()

	a.metrics["Alloc"] = float64(memStats.Alloc)
	a.metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	a.metrics["Frees"] = float64(memStats.Frees)
	a.metrics["GCCPUFraction"] = memStats.GCCPUFraction
	a.metrics["GCSys"] = float64(memStats.GCSys)
	a.metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	a.metrics["HeapIdle"] = float64(memStats.HeapIdle)
	a.metrics["HeapInuse"] = float64(memStats.HeapInuse)
	a.metrics["HeapObjects"] = float64(memStats.HeapObjects)
	a.metrics["HeapReleased"] = float64(memStats.HeapReleased)
	a.metrics["HeapSys"] = float64(memStats.HeapSys)
	a.metrics["LastGC"] = float64(memStats.LastGC)
	a.metrics["Lookups"] = float64(memStats.Lookups)
	a.metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	a.metrics["MCacheSys"] = float64(memStats.MCacheSys)
	a.metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	a.metrics["MSpanSys"] = float64(memStats.MSpanSys)
	a.metrics["Mallocs"] = float64(memStats.Mallocs)
	a.metrics["NextGC"] = float64(memStats.NextGC)
	a.metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	a.metrics["NumGC"] = float64(memStats.NumGC)
	a.metrics["OtherSys"] = float64(memStats.OtherSys)
	a.metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	a.metrics["StackInuse"] = float64(memStats.StackInuse)
	a.metrics["StackSys"] = float64(memStats.StackSys)
	a.metrics["Sys"] = float64(memStats.Sys)
	a.metrics["TotalAlloc"] = float64(memStats.TotalAlloc)
}

func (a *Agent) send() {
	a.logger.Info("Metrics sending...")
	start := time.Now()
	batch := buildBatch(a.metrics)
	if len(batch) <= 0 {
		return
	}

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
