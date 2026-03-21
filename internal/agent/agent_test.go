package agent

import "testing"

func TestPoll_AllMetricsPresent(t *testing.T) {
	metrics := make(Metrics)
	Poll(metrics, 1)

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
