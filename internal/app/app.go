// Package app инициализирует и связывает все компоненты приложения.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"metrics-collector/internal/audit"
	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
	"metrics-collector/internal/worker"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// App управляет жизненным циклом сервера.
type App struct {
	cfg            *config.ServerConfig
	logger         *zap.Logger
	server         *http.Server
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

	r := registerRoutes(healthHandler, metricsHandler, auditPublisher, signer, enc, logger, cfg)

	server := &http.Server{
		Addr:         cfg.ServerBaseURL,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return &App{
		cfg:            cfg,
		logger:         logger,
		server:         server,
		auditPublisher: auditPublisher,
		backupWorker:   backupWorker,
		backupRepo:     repo,
	}, nil
}

// Run запускает сервер и ожидает сигнала завершения.
func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		a.logger.Info("Server started", zap.String("url", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

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
		return a.shutdown()
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		return err
	}

	return nil
}

func (a *App) shutdown() error {
	a.logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()

	if err := a.backupRepo.BackupMetrics(shutdownCtx); err != nil {
		a.logger.Error("failed to backup metrics during shutdown", zap.Error(err))
	}

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	a.logger.Info("Server stopped gracefully")
	return nil
}
