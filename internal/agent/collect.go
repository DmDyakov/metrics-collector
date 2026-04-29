package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

func (a *Agent) collectMemStats(count int64) {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	a.store.mu.Lock()

	a.store.metrics["PollCount"] = float64(count)
	a.store.metrics["RandomValue"] = rand.Float64()

	a.store.metrics["Alloc"] = float64(memStats.Alloc)
	a.store.metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	a.store.metrics["Frees"] = float64(memStats.Frees)
	a.store.metrics["GCCPUFraction"] = memStats.GCCPUFraction
	a.store.metrics["GCSys"] = float64(memStats.GCSys)
	a.store.metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	a.store.metrics["HeapIdle"] = float64(memStats.HeapIdle)
	a.store.metrics["HeapInuse"] = float64(memStats.HeapInuse)
	a.store.metrics["HeapObjects"] = float64(memStats.HeapObjects)
	a.store.metrics["HeapReleased"] = float64(memStats.HeapReleased)
	a.store.metrics["HeapSys"] = float64(memStats.HeapSys)
	a.store.metrics["LastGC"] = float64(memStats.LastGC)
	a.store.metrics["Lookups"] = float64(memStats.Lookups)
	a.store.metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	a.store.metrics["MCacheSys"] = float64(memStats.MCacheSys)
	a.store.metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	a.store.metrics["MSpanSys"] = float64(memStats.MSpanSys)
	a.store.metrics["Mallocs"] = float64(memStats.Mallocs)
	a.store.metrics["NextGC"] = float64(memStats.NextGC)
	a.store.metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	a.store.metrics["NumGC"] = float64(memStats.NumGC)
	a.store.metrics["OtherSys"] = float64(memStats.OtherSys)
	a.store.metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	a.store.metrics["StackInuse"] = float64(memStats.StackInuse)
	a.store.metrics["StackSys"] = float64(memStats.StackSys)
	a.store.metrics["Sys"] = float64(memStats.Sys)
	a.store.metrics["TotalAlloc"] = float64(memStats.TotalAlloc)

	a.store.mu.Unlock()
}

func (a *Agent) collectVirtualMemoryInfo() {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		a.logger.Warn("failed to get memory metrics", zap.Error(err))
		return
	}

	a.store.mu.Lock()
	a.store.metrics["TotalMemory"] = float64(memInfo.Total)
	a.store.metrics["FreeMemory"] = float64(memInfo.Free)
	a.store.mu.Unlock()

}

func (a *Agent) collectCpuPercentsInfo() {
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		a.logger.Warn("failed to get cpu percent", zap.Error(err))
		return
	}

	a.store.mu.Lock()
	for i, v := range cpuPercents {
		key := fmt.Sprintf("CPUutilization%d", i+1)
		a.store.metrics[key] = v
	}
	a.store.mu.Unlock()

}
