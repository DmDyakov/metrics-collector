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
	"metrics-collector/internal/server/repository/file"
	"metrics-collector/internal/server/repository/mem"
	"metrics-collector/internal/server/repository/postgres"
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
	repo           *repository.Repository
	auditPublisher *audit.Publisher
	backupWorker   *worker.BackupWorker
}

// New создаёт новый App.
func New(cfg *config.ServerConfig, logger *zap.Logger) (*App, error) {
	// Хранилища
	memStorage := mem.NewMemStorage()

	var persistent repository.PersistentStorage
	if cfg.DatabaseDSN != "" {
		pg, _, err := postgres.NewPostgresStorage(cfg.DatabaseDSN, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres storage: %w", err)
		}
		persistent = pg
		logger.Info("Storage mode: postgres")
	} else if cfg.FileStoragePath != "" {
		persistent = file.NewFileStorage(cfg.FileStoragePath)
		logger.Info("Storage mode: file")
	} else {
		logger.Info("Storage mode: memory only")
	}

	repo := repository.New(memStorage, persistent, cfg.StoreInterval, logger)

	// Воркеры
	backupWorker := worker.NewBackupWorker(
		cfg.Restore,
		cfg.StoreInterval,
		repo,
		logger,
	)

	// Сервисы
	metricsService := service.NewMetricsService(repo)
	healthService := service.NewHealthService(repo)

	// HTTP handler'ы
	metricsHandler, err := handler.NewMetricsHandler(metricsService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics handler: %w", err)
	}
	healthHandler := handler.NewHealthHandler(healthService, logger)

	// Аудит
	auditPublisher, err := audit.NewPublisher(cfg.AuditFile, cfg.AuditURL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit publisher: %w", err)
	}

	// Подпись
	signer, err := signer.New(cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signer: %w", err)
	}
	enc, err := encryptor.New("", cfg.PrivateCryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// HTTP сервер
	httpServer := httpserver.New(
		cfg,
		healthHandler,
		metricsHandler,
		auditPublisher,
		signer,
		enc,
		logger,
	)

	// gRPC сервер
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
		repo:           repo,
		auditPublisher: auditPublisher,
		backupWorker:   backupWorker,
	}, nil
}

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

	g.Go(func() error {
		<-ctx.Done()
		a.repo.Close()
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
