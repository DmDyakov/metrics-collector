// Package worker содержит воркеры агента.
package worker

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	"metrics-collector/internal/domain/metrics"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_poller_store.go -package=mocks . PollerStore
type PollerStore interface {
	UpdateMetrics(metrics map[string]float64)
}

// Poller периодически собирает метрики системы и сохраняет в хранилище.
type Poller struct {
	store        PollerStore
	logger       *zap.Logger
	pollInterval time.Duration
}

// NewPoller создаёт Poller.
func NewPoller(s PollerStore, l *zap.Logger, pollInterval int) *Poller {
	return &Poller{
		store:        s,
		logger:       l,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}
}

// Run запускает периодический сбор метрик.
func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()
	var count int64
	p.logger.Info("Poller started", zap.Duration("interval", p.pollInterval))

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Poller stopped")
			return nil
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

// collectMemStats собирает метрики runtime.MemStats.
func (p *Poller) pollMemStats(count int64) {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	p.store.UpdateMetrics(map[string]float64{
		metrics.PollCount:     float64(count),
		metrics.RandomValue:   rand.Float64(),
		metrics.Alloc:         float64(memStats.Alloc),
		metrics.BuckHashSys:   float64(memStats.BuckHashSys),
		metrics.Frees:         float64(memStats.Frees),
		metrics.GCCPUFraction: memStats.GCCPUFraction,
		metrics.GCSys:         float64(memStats.GCSys),
		metrics.HeapAlloc:     float64(memStats.HeapAlloc),
		metrics.HeapIdle:      float64(memStats.HeapIdle),
		metrics.HeapInuse:     float64(memStats.HeapInuse),
		metrics.HeapObjects:   float64(memStats.HeapObjects),
		metrics.HeapReleased:  float64(memStats.HeapReleased),
		metrics.HeapSys:       float64(memStats.HeapSys),
		metrics.LastGC:        float64(memStats.LastGC),
		metrics.Lookups:       float64(memStats.Lookups),
		metrics.MCacheInuse:   float64(memStats.MCacheInuse),
		metrics.MCacheSys:     float64(memStats.MCacheSys),
		metrics.MSpanInuse:    float64(memStats.MSpanInuse),
		metrics.MSpanSys:      float64(memStats.MSpanSys),
		metrics.Mallocs:       float64(memStats.Mallocs),
		metrics.NextGC:        float64(memStats.NextGC),
		metrics.NumForcedGC:   float64(memStats.NumForcedGC),
		metrics.NumGC:         float64(memStats.NumGC),
		metrics.OtherSys:      float64(memStats.OtherSys),
		metrics.PauseTotalNs:  float64(memStats.PauseTotalNs),
		metrics.StackInuse:    float64(memStats.StackInuse),
		metrics.StackSys:      float64(memStats.StackSys),
		metrics.Sys:           float64(memStats.Sys),
		metrics.TotalAlloc:    float64(memStats.TotalAlloc),
	})
}

// collectVirtualMemory собирает метрики виртуальной памяти.
func (p *Poller) pollVirtualMemoryInfo() {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		p.logger.Warn("Failed to get memory metrics", zap.Error(err))
		return
	}

	p.store.UpdateMetrics(map[string]float64{
		metrics.TotalMemory: float64(memInfo.Total),
		metrics.FreeMemory:  float64(memInfo.Free),
	})
}

// collectCPUPercent собирает метрики загрузки CPU.
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
