package agent

import (
	"context"
	"net/http"
	"sync"
	"time"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"

	"go.uber.org/zap"
)

const (
	PollCount   = "PoolCount"
	RandomValue = "RandomValue"
)

type Store struct {
	metrics storeMetrics
	mu      sync.RWMutex
}

type storeMetrics map[string]float64

type batchMetric struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"` // для Counter
	Value *float64 `json:"value,omitempty"` // для Gauge
}

type Agent struct {
	cfg    *config.AgentConfig
	logger *zap.Logger
	gzip   *compress.Gzip
	client *http.Client
	store  *Store

	jobs chan []batchMetric
}

func NewStore() *Store {
	return &Store{
		metrics: make(map[string]float64),
	}
}

func NewAgent(cfg *config.AgentConfig, logger *zap.Logger, gzip *compress.Gzip) *Agent {
	return &Agent{
		cfg:    cfg,
		logger: logger,
		gzip:   gzip,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		store: NewStore(),
		jobs:  make(chan []batchMetric, cfg.RateLimit*2), // *2 - чтобы сгладить пики и избежать блокировки reportScheduler
	}
}

func (a *Agent) Run(ctx context.Context) {
	a.logger.Info("Starting agent",
		zap.Int("rate_limit", a.cfg.RateLimit),
		zap.Int("poll_interval", a.cfg.PollInterval),
		zap.Int("report_interval", a.cfg.ReportInterval),
	)

	var wg sync.WaitGroup

	for i := 0; i < a.cfg.RateLimit; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a.reportingWorker(ctx, id)
		}(i + 1)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.collectRuntimeMetrics(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.collectSystemMetrics(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.reportingScheduler(ctx)
	}()

	<-ctx.Done()

	close(a.jobs)

	wg.Wait()
	a.logger.Info("Agent stopped")
}

func (a *Agent) reportingWorker(ctx context.Context, id int) {
	for job := range a.jobs {
		a.logger.Debug("worker started",
			zap.Int("id", id),
		)

		a.reportBatch(job)

		a.logger.Debug("worker finished",
			zap.Int("id", id),
		)
	}
	a.logger.Debug("reporting worker stopped", zap.Int("id", id))
}

func (a *Agent) reportingScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Debug("reporting scheduler stopped, context done")
			return
		case <-ticker.C:
			a.logger.Info("schedule reporting")
			batch := a.buildReportingBatch()

			select {
			case <-ctx.Done():
				a.logger.Debug("report batch enqueued skipped, reporting scheduler stopped, context cancelled")
				return
			case a.jobs <- batch:
				a.logger.Debug("report batch enqueued")
			}
		}
	}

}

func (a *Agent) collectRuntimeMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	defer ticker.Stop()
	var count int64

	for {
		select {
		case <-ctx.Done():
			a.logger.Debug("runtime metrics collector stopped, context done")
			return
		case <-ticker.C:
			count++
			a.logger.Info("Collect runtime metrics",
				zap.Int64("iteration", count),
			)

			a.collectMemStats(count)
		}
	}
}

func (a *Agent) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Debug("system metrics collector stopped, context done")
			return
		case <-ticker.C:
			a.logger.Info("Collect system metrics")
			a.collectVirtualMemoryInfo()
			a.collectCpuPercentsInfo()
		}
	}
}
