// Package app собирает и запускает все серверы.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"metrics-collector/internal/config"
	"metrics-collector/internal/server/audit"
	"metrics-collector/internal/server/repository"
	"metrics-collector/internal/server/service"
	grpcserver "metrics-collector/internal/server/transport/grpc"
	httpserver "metrics-collector/internal/server/transport/http"
	"metrics-collector/internal/server/transport/http/handler"
	"metrics-collector/internal/server/worker"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// App управляет жизненным циклом сервера.
type App struct {
	cfg            *config.ServerConfig
	logger         *zap.Logger
	httpServer     *http.Server
	grpcServer     *grpc.Server
	auditPublisher *audit.Publisher
	backupWorker   *worker.BackupWorker
	backupRepo     worker.BackupRepository
}

// New создаёт новый App.
func New(cfg *config.ServerConfig, logger *zap.Logger) (*App, error) {
	repo, err := repository.NewRepository(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	backupWorker := worker.NewBackupWorker(
		cfg.Restore,
		cfg.StoreInterval,
		repo,
		logger,
	)

	metricsService := service.NewMetricsService(repo)
	healthService := service.NewHealthService(repo)

	metricsHandler, err := handler.NewMetricsHandler(metricsService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics handler: %w", err)
	}
	healthHandler := handler.NewHealthHandler(healthService, logger)

	auditPublisher, err := audit.NewPublisher(cfg.AuditFile, cfg.AuditURL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit publisher: %w", err)
	}

	signer, err := signer.New(cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signer: %w", err)
	}

	enc, err := encryptor.New("", cfg.PrivateCryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	httpServer := httpserver.New(
		cfg,
		healthHandler,
		metricsHandler,
		auditPublisher,
		signer,
		enc,
		logger,
	)

	var grpcServer *grpc.Server
	if cfg.GRPCAddress != "" {
		grpcHandler := grpcserver.NewHandler(metricsService, logger)
		grpcServer, err = grpcserver.New(cfg.GRPCAddress, cfg.TrustedSubnet, grpcHandler, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC server: %w", err)
		}
	}

	return &App{
		cfg:            cfg,
		logger:         logger,
		httpServer:     httpServer,
		grpcServer:     grpcServer,
		auditPublisher: auditPublisher,
		backupWorker:   backupWorker,
		backupRepo:     repo,
	}, nil
}

// Run запускает все серверы и воркеры.
func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return httpserver.Run(ctx, a.httpServer, a.cfg.ShutdownTimeout, a.logger)
	})

	if a.grpcServer != nil {
		g.Go(func() error {
			return grpcserver.Run(ctx, a.grpcServer, a.logger)
		})
	}

	g.Go(func() error {
		return a.backupWorker.Run(ctx)
	})

	if a.auditPublisher != nil {
		g.Go(func() error {
			return a.auditPublisher.Run(ctx)
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
