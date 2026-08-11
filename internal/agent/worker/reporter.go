package worker

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -destination=mocks/mock_reporter_store.go -package=mocks . ReporterStore
type ReporterStore interface {
	UpdateMetrics(metrics map[string]float64)
	GetMetricsSnapshot() map[string]float64
}

//go:generate mockgen -destination=mocks/mock_client.go -package=mocks . Client
type Client interface {
	SendMetrics(ctx context.Context, metrics map[string]float64) error
	io.Closer
}

// Reporter периодически отправляет метрики на сервер через пул воркеров.
type Reporter struct {
	client Client
	store  ReporterStore
	logger *zap.Logger

	reportInterval int
	jobs           chan map[string]float64
	workers        int
}

// NewReporter создаёт Reporter.
func NewReporter(s ReporterStore, c Client, l *zap.Logger, workers int, reportInterval int) *Reporter {
	return &Reporter{
		client:         c,
		store:          s,
		logger:         l,
		reportInterval: reportInterval,
		jobs:           make(chan map[string]float64, workers*2),
		workers:        workers,
	}
}

// Run запускает пул воркеров и планировщик отправки метрик.
func (r *Reporter) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < r.workers; i++ {
		workerID := i + 1
		g.Go(func() error {
			return r.worker(ctx, workerID)
		})
	}

	g.Go(func() error {
		return r.scheduler(ctx)
	})

	r.logger.Info("Reporter pool started",
		zap.Int("workers", r.workers),
		zap.Duration("interval", time.Duration(r.reportInterval)*time.Second),
	)

	if err := g.Wait(); err != nil && err != context.Canceled {
		return fmt.Errorf("reporter: %w", err)
	}

	return nil
}

// worker отправляет метрики из очереди на сервер.
func (r *Reporter) worker(ctx context.Context, id int) error {
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Reporter stopped", zap.Int("worker_id", id))
			return nil
		case batch, ok := <-r.jobs:
			if !ok {
				return nil
			}
			if err := r.client.SendMetrics(ctx, batch); err != nil {
				r.logger.Warn("Failed to send metrics",
					zap.Int("worker_id", id),
					zap.Error(err),
				)
			}
		}
	}
}

// scheduler периодически снимает снапшот метрик и отправляет в очередь.
func (r *Reporter) scheduler(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(r.reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Reporter scheduler stopped")
			return ctx.Err()
		case <-ticker.C:
			batch := r.store.GetMetricsSnapshot()
			select {
			case r.jobs <- batch:
				r.logger.Debug("Batch enqueued")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
