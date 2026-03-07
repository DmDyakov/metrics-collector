package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"metrics-collector/internal/config"
)

type Metrics map[string]string

func Run(cfg *config.Config) {
	var metrics = make(Metrics)
	reportMultiplier := int64(cfg.ReportInterval / cfg.PollInterval)
	var pollCount int64 = 0

	for {

		pollCount++

		Poll(metrics, pollCount)
		time.Sleep(cfg.PollInterval)

		if pollCount%reportMultiplier == 0 {
			Send(metrics, cfg.ServerBaseURL)
		}

	}

}

func Poll(metrics Metrics, count int64) {
	log.Printf("--- Опрос метрик #%d ---\n", count)

	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	metrics["PollCount"] = fmt.Sprintf("%d", count)
	metrics["RandomValue"] = fmt.Sprintf("%f", rand.Float64())

	metrics["Alloc"] = fmt.Sprintf("%f", float64(memStats.Alloc))
	metrics["BuckHashSys"] = fmt.Sprintf("%f", float64(memStats.BuckHashSys))
	metrics["Frees"] = fmt.Sprintf("%f", float64(memStats.Frees))
	metrics["GCCPUFraction"] = fmt.Sprintf("%f", memStats.GCCPUFraction)
	metrics["GCSys"] = fmt.Sprintf("%f", float64(memStats.GCSys))
	metrics["HeapAlloc"] = fmt.Sprintf("%f", float64(memStats.HeapAlloc))
	metrics["HeapIdle"] = fmt.Sprintf("%f", float64(memStats.HeapIdle))
	metrics["HeapInuse"] = fmt.Sprintf("%f", float64(memStats.HeapInuse))
	metrics["HeapObjects"] = fmt.Sprintf("%f", float64(memStats.HeapObjects))
	metrics["HeapReleased"] = fmt.Sprintf("%f", float64(memStats.HeapReleased))
	metrics["HeapSys"] = fmt.Sprintf("%f", float64(memStats.HeapSys))
	metrics["LastGC"] = fmt.Sprintf("%f", float64(memStats.LastGC))
	metrics["Lookups"] = fmt.Sprintf("%f", float64(memStats.Lookups))
	metrics["MCacheInuse"] = fmt.Sprintf("%f", float64(memStats.MCacheInuse))
	metrics["MCacheSys"] = fmt.Sprintf("%f", float64(memStats.MCacheSys))
	metrics["MSpanInuse"] = fmt.Sprintf("%f", float64(memStats.MSpanInuse))
	metrics["MSpanSys"] = fmt.Sprintf("%f", float64(memStats.MSpanSys))
	metrics["Mallocs"] = fmt.Sprintf("%f", float64(memStats.Mallocs))
	metrics["NextGC"] = fmt.Sprintf("%f", float64(memStats.NextGC))
	metrics["NumForcedGC"] = fmt.Sprintf("%f", float64(memStats.NumForcedGC))
	metrics["NumGC"] = fmt.Sprintf("%f", float64(memStats.NumGC))
	metrics["OtherSys"] = fmt.Sprintf("%f", float64(memStats.OtherSys))
	metrics["PauseTotalNs"] = fmt.Sprintf("%f", float64(memStats.PauseTotalNs))
	metrics["StackInuse"] = fmt.Sprintf("%f", float64(memStats.StackInuse))
	metrics["StackSys"] = fmt.Sprintf("%f", float64(memStats.StackSys))
	metrics["Sys"] = fmt.Sprintf("%f", float64(memStats.Sys))
	metrics["TotalAlloc"] = fmt.Sprintf("%f", float64(memStats.TotalAlloc))
}

func Send(metrics Metrics, baseURL string) {
	log.Println("--- Отправка метрик ---")
	for name, value := range metrics {

		metricType := "gauge"
		if name == "PollCount" {
			metricType = "counter"
		}

		url := fmt.Sprintf("http://%s/update/%s/%s/%s", baseURL, metricType, name, value)
		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp.Body.Close()

		fmt.Printf("Response status: %s\n", resp.Status)

	}

}
