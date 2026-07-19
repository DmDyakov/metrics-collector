// Package worker содержит воркеры агента.
package worker

import (
	"context"
	"fmt"
	"math/rand/v2"
	models "metrics-collector/internal/model"
	"time"

	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_poller_store.go -package=mocks . PollerStore
type PollerStore interface {
	UpdateMetrics(metrics map[string]float64)
}

// generate:reset
type Poller struct {
	store        PollerStore
	logger       *zap.Logger
	pollInterval int
}

func NewPoller(s PollerStore, l *zap.Logger, pollInterval int) *Poller {
	return &Poller{
		store:        s,
		logger:       l,
		pollInterval: pollInterval,
	}
}

func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(p.pollInterval) * time.Second)
	defer ticker.Stop()
	var count int64

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Poller stopped")
			return ctx.Err()
		case <-ticker.C:
			count++
			p.logger.Info("Poll runtime metrics",
				zap.Int64("iteration", count),
			)

			p.pollMemStats(count)
			p.pollVirtualMemoryInfo()
			p.pollCPUPercentsInfo()
		}
	}
}

func (p *Poller) pollMemStats(count int64) {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	p.store.UpdateMetrics(map[string]float64{
		models.PollCount:   float64(count),
		models.RandomValue: rand.Float64(),
		"Alloc":            float64(memStats.Alloc),
		"BuckHashSys":      float64(memStats.BuckHashSys),
		"Frees":            float64(memStats.Frees),
		"GCCPUFraction":    memStats.GCCPUFraction,
		"GCSys":            float64(memStats.GCSys),
		"HeapAlloc":        float64(memStats.HeapAlloc),
		"HeapIdle":         float64(memStats.HeapIdle),
		"HeapInuse":        float64(memStats.HeapInuse),
		"HeapObjects":      float64(memStats.HeapObjects),
		"HeapReleased":     float64(memStats.HeapReleased),
		"HeapSys":          float64(memStats.HeapSys),
		"LastGC":           float64(memStats.LastGC),
		"Lookups":          float64(memStats.Lookups),
		"MCacheInuse":      float64(memStats.MCacheInuse),
		"MCacheSys":        float64(memStats.MCacheSys),
		"MSpanInuse":       float64(memStats.MSpanInuse),
		"MSpanSys":         float64(memStats.MSpanSys),
		"Mallocs":          float64(memStats.Mallocs),
		"NextGC":           float64(memStats.NextGC),
		"NumForcedGC":      float64(memStats.NumForcedGC),
		"NumGC":            float64(memStats.NumGC),
		"OtherSys":         float64(memStats.OtherSys),
		"PauseTotalNs":     float64(memStats.PauseTotalNs),
		"StackInuse":       float64(memStats.StackInuse),
		"StackSys":         float64(memStats.StackSys),
		"Sys":              float64(memStats.Sys),
		"TotalAlloc":       float64(memStats.TotalAlloc),
	})

}

func (p *Poller) pollVirtualMemoryInfo() {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		p.logger.Warn("Failed to get memory metrics", zap.Error(err))
		return
	}

	p.store.UpdateMetrics(map[string]float64{
		"TotalMemory": float64(memInfo.Total),
		"FreeMemory":  float64(memInfo.Free),
	})
}

func (p *Poller) pollCPUPercentsInfo() {
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		p.logger.Warn("Failed to get cpu percent", zap.Error(err))
		return
	}

	batch := make(map[string]float64)

	for i, v := range cpuPercents {
		key := fmt.Sprintf("CPUutilization%d", i+1)
		batch[key] = v
	}

	p.store.UpdateMetrics(batch)
}
