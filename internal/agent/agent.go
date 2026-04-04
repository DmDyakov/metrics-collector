package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"
	models "metrics-collector/internal/model"

	"go.uber.org/zap"
)

type Metrics map[string]any

type Agent struct {
	cfg    *config.AgentConfig
	logger *zap.Logger
	gzip   *compress.Gzip
	client *http.Client
}

func NewAgent(cfg *config.AgentConfig, logger *zap.Logger, gzip *compress.Gzip) *Agent {
	return &Agent{
		cfg:    cfg,
		logger: logger,
		gzip:   gzip,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (a *Agent) Run() {
	var metrics = make(Metrics)
	reportMultiplier := int64(a.cfg.ReportInterval / a.cfg.PollInterval)
	var pollCount int64 = 0

	pollTicker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()

	for range pollTicker.C {
		pollCount++

		a.Poll(metrics, pollCount)

		if pollCount%reportMultiplier == 0 {
			a.Send(metrics)
		}
	}

}

func (a *Agent) Poll(metrics Metrics, count int64) {
	a.logger.Info("Опрос метрик",
		zap.Int64("iteration", count),
	)

	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	metrics["PollCount"] = count
	metrics["RandomValue"] = rand.Float64()

	metrics["Alloc"] = float64(memStats.Alloc)
	metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	metrics["Frees"] = float64(memStats.Frees)
	metrics["GCCPUFraction"] = memStats.GCCPUFraction
	metrics["GCSys"] = float64(memStats.GCSys)
	metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	metrics["HeapIdle"] = float64(memStats.HeapIdle)
	metrics["HeapInuse"] = float64(memStats.HeapInuse)
	metrics["HeapObjects"] = float64(memStats.HeapObjects)
	metrics["HeapReleased"] = float64(memStats.HeapReleased)
	metrics["HeapSys"] = float64(memStats.HeapSys)
	metrics["LastGC"] = float64(memStats.LastGC)
	metrics["Lookups"] = float64(memStats.Lookups)
	metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	metrics["MCacheSys"] = float64(memStats.MCacheSys)
	metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	metrics["MSpanSys"] = float64(memStats.MSpanSys)
	metrics["Mallocs"] = float64(memStats.Mallocs)
	metrics["NextGC"] = float64(memStats.NextGC)
	metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	metrics["NumGC"] = float64(memStats.NumGC)
	metrics["OtherSys"] = float64(memStats.OtherSys)
	metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	metrics["StackInuse"] = float64(memStats.StackInuse)
	metrics["StackSys"] = float64(memStats.StackSys)
	metrics["Sys"] = float64(memStats.Sys)
	metrics["TotalAlloc"] = float64(memStats.TotalAlloc)
}

func (a *Agent) Send(metrics Metrics) {
	a.logger.Info("--- Отправка метрик ---")
	for name, value := range metrics {
		var payload Metrics
		payload = Metrics{
			"id":    name,
			"type":  models.Gauge,
			"value": value,
		}

		if name == "PollCount" {
			payload = Metrics{
				"id":    name,
				"type":  models.Counter,
				"delta": value,
			}
		}

		start := time.Now()

		url := fmt.Sprintf("http://%s/update", a.cfg.ServerBaseURL)

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			a.logger.Error("Error JSON marshaling", zap.Error(err))
			continue
		}

		compressedJson, err := a.gzip.Compress(jsonPayload)
		if err != nil {
			a.logger.Error("Error JSON compressing", zap.Error(err))
			continue
		}

		method := "POST"
		req, err := http.NewRequest(method, url, bytes.NewReader(compressedJson))
		if err != nil {
			a.logger.Error("failed to build request",
				zap.String("uri", url),
				zap.String("method", method),
				zap.Error(err),
			)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := a.client.Do(req)
		if err != nil {
			a.logger.Error("failed to send request",
				zap.String("uri", url),
				zap.String("method", method),
				zap.Error(err),
			)
			continue
		}
		resp.Body.Close()

		duration := time.Since(start)
		a.logger.Info("request sent",
			zap.String("uri", url),
			zap.String("method", "POST"),
			zap.Duration("duration", duration),
		)
	}
}
