package worker

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Reporter struct {
	client Client
	store  Store
	logger *zap.Logger

	reportInterval int
	jobs           chan map[string]float64
	workers        int
}

func NewReporter(s Store, c Client, l *zap.Logger, workers int, reportInterval int) *Reporter {
	return &Reporter{
		client:         c,
		store:          s,
		logger:         l,
		reportInterval: reportInterval,
		jobs:           make(chan map[string]float64, workers*2),
		workers:        workers,
	}
}

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
