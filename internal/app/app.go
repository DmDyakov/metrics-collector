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

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// App управляет жизненным циклом сервера.
type App struct {
	cfg            *config.ServerConfig
	logger         *zap.Logger
	server         *http.Server
	auditPublisher *audit.Publisher
}

// New создаёт новый App.
func New(cfg *config.ServerConfig, logger *zap.Logger) (*App, error) {
	repo, err := repository.NewRepository(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

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

	r := registerRoutes(healthHandler, metricsHandler, auditPublisher, logger, cfg)

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

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	a.logger.Info("Server stopped gracefully")
	return nil
}
