// Package agent реализует агента сбора метрик
package agent

import (
	"context"

	"metrics-collector/internal/agent/client"
	"metrics-collector/internal/agent/store"
	"metrics-collector/internal/agent/worker"
	"metrics-collector/internal/config"
	"metrics-collector/pkg/compress"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Agent struct {
	cfg      *config.AgentConfig
	logger   *zap.Logger
	gzip     *compress.Gzip
	store    *store.Store
	poller   *worker.Poller
	reporter *worker.Reporter
}

// NewAgent creates a new agent with the given configuration.
func NewAgent(cfg *config.AgentConfig, l *zap.Logger) *Agent {
	store := store.New()
	gzip := compress.NewGzip()
	signer, err := signer.New(cfg.SecretKey)
	if err != nil {
		l.Error(err.Error())
	}
	encryptor, err := encryptor.New(cfg.PublicCryptoKey, "")
	if err != nil {
		l.Fatal("failed to load public key", zap.Error(err))
	}
	client := client.New(cfg.ServerBaseURL, cfg.AgentIP, l, signer, encryptor, gzip)
	poller := worker.NewPoller(store, l, cfg.PollInterval)
	reporter := worker.NewReporter(store, client, l, cfg.RateLimit, cfg.ReportInterval)

	return &Agent{
		cfg:      cfg,
		logger:   l,
		gzip:     gzip,
		store:    store,
		poller:   poller,
		reporter: reporter,
	}
}

// Run starts the agent: metrics collection and reporting to the server.
func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info("Starting agent",
		zap.Int("rate_limit", a.cfg.RateLimit),
		zap.Int("poll_interval", a.cfg.PollInterval),
		zap.Int("report_interval", a.cfg.ReportInterval),
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		a.logger.Info("Starting poller")
		return a.poller.Run(ctx)
	})

	g.Go(func() error {
		a.logger.Info("Starting reporter")
		return a.reporter.Run(ctx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		return err
	}

	return nil
}
