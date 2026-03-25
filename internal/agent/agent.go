package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"metrics-collector/internal/config"

	"go.uber.org/zap"
)

type Metrics map[string]any

func Run(cfg *config.AgentConfig, logger *zap.SugaredLogger) {
	var metrics = make(Metrics)
	reportMultiplier := int64(cfg.ReportInterval / cfg.PollInterval)
	var pollCount int64 = 0

	for {

		pollCount++

		Poll(metrics, pollCount, logger)
		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)

		if pollCount%reportMultiplier == 0 {
			Send(metrics, cfg.ServerBaseURL, logger)
		}

	}

}

func Poll(metrics Metrics, count int64, logger *zap.SugaredLogger) {
	logger.Infof("--- Опрос метрик #%d ---", count)

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

func Send(metrics Metrics, baseURL string, logger *zap.SugaredLogger) {
	logger.Infoln("--- Отправка метрик ---")
	for name, value := range metrics {
		var payload Metrics
		payload = Metrics{
			"id":    name,
			"type":  "gauge",
			"value": value,
		}

		if name == "PollCount" {
			payload = Metrics{
				"id":    name,
				"type":  "counter",
				"delta": value,
			}
		}

		start := time.Now()

		url := fmt.Sprintf("http://%s/update", baseURL)

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			return
		}
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			logger.Errorw("ошибка отправки запроса",
				"uri", url,
				"method", "POST",
				"error", err,
			)
			continue
		}
		resp.Body.Close()

		duration := time.Since(start)
		logger.Infoln(
			"uri", url,
			"method", "POST",
			"duration", duration,
		)
	}
}
