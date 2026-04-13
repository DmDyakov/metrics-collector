package agent

import (
	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"
	"testing"

	"go.uber.org/zap"
)

func TestAgent_Poll(t *testing.T) {
	metrics := make(Metrics)
	logger := zap.NewNop()
	gzip := compress.NewGzip()
	cfg := &config.AgentConfig{
		PollInterval:   2,
		ReportInterval: 10,
		ServerBaseURL:  "localhost:8080",
	}
	a := NewAgent(cfg, logger, gzip)
	a.Poll(metrics, 1)

	expectedMetrics := []string{
		"PollCount",
		"RandomValue",
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}

	for _, name := range expectedMetrics {
		if _, ok := metrics[name]; !ok {
			t.Errorf("metric %q not found after Poll()", name)
		}
	}
}
