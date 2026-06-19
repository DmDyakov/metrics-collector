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
	"metrics-collector/internal/middleware"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
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

	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.WithTimeout(cfg.RequestTimeout))
	r.Use(middleware.WithLogging(logger))
	r.Use(middleware.WithSignature(logger, cfg.SecretKey))
	r.Use(middleware.WithCompressing)

	r.Get("/ping", healthHandler.HealthDB)

	r.Group(func(r chi.Router) {
		r.Get("/", metricsHandler.ListMetrics)
		r.Get("/value/{type}/{name}", metricsHandler.GetMetricValue)
		r.Post("/value", metricsHandler.GetMetric)
	})

	r.Group(func(r chi.Router) {
		if auditPublisher != nil {
			r.Use(middleware.WithAudit(auditPublisher))
		}

		r.Post("/update/{type}/{name}/{value}", metricsHandler.UpdateMetricByURL)
		r.Post("/update", metricsHandler.UpdateMetricByJSON)
		r.Post("/updates", metricsHandler.UpdateMetricsBatch)
	})

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
	appCtx, appCancel := context.WithCancel(ctx)
	defer appCancel()

	go func() {
		a.logger.Info("Server started", zap.String("url", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("listen error", zap.Error(err))
			appCancel()
		}
	}()

	<-appCtx.Done()
	a.logger.Info("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if a.auditPublisher != nil {
		auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer auditCancel()
		if err := a.auditPublisher.Shutdown(auditCtx); err != nil {
			a.logger.Warn("audit shutdown failed", zap.Error(err))
		} else {
			a.logger.Info("Audit publisher stopped gracefully")
		}
	}

	a.logger.Info("Server stopped gracefully")
	return nil
}
