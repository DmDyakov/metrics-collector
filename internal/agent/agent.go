// Package agent реализует агента сбора метрик
package agent

import (
	"context"
	"fmt"
	"io"
	grpcclient "metrics-collector/internal/agent/clients/grpc"
	httpclient "metrics-collector/internal/agent/clients/http"
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
	client   io.Closer
	store    *store.Store
	poller   *worker.Poller
	reporter *worker.Reporter
}

// New creates a new agent with the given configuration.
func New(cfg *config.AgentConfig, logger *zap.Logger) (*Agent, error) {
	client, err := newClient(cfg, logger)
	if err != nil {
		return nil, err
	}

	store := store.New()
	poller := worker.NewPoller(store, logger, cfg.PollInterval)
	reporter := worker.NewReporter(store, client, logger, cfg.RateLimit, cfg.ReportInterval)

	return &Agent{
		cfg:      cfg,
		logger:   logger,
		client:   client,
		store:    store,
		poller:   poller,
		reporter: reporter,
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	defer a.client.Close()

	a.logger.Info("Starting agent",
		zap.Int("rate_limit", a.cfg.RateLimit),
		zap.Int("poll_interval", a.cfg.PollInterval),
		zap.Int("report_interval", a.cfg.ReportInterval),
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.poller.Run(ctx)
	})

	g.Go(func() error {
		return a.reporter.Run(ctx)
	})

	return g.Wait()
}

func newClient(cfg *config.AgentConfig, logger *zap.Logger) (worker.Client, error) {
	switch {
	case cfg.GRPCAddress != "":
		client, err := grpcclient.New(cfg, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC client: %w", err)
		}
		logger.Info("Using gRPC client", zap.String("addr", cfg.GRPCAddress))
		return client, nil

	default:
		gzip := compress.NewGzip()

		signer, err := signer.New(cfg.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create signer: %w", err)
		}

		encryptor, err := encryptor.New(cfg.PublicCryptoKey, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}

		client := httpclient.New(cfg.HTTPAddress, cfg.AgentIP, logger, signer, encryptor, gzip)
		logger.Info("Using HTTP client", zap.String("addr", cfg.HTTPAddress))
		return client, nil
	}
}
