package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/logger"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"

	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("server app failed: %v", err)
	}
}

func run(ctx context.Context) error {
	logger, err := logger.NewZapLogger()
	if err != nil {
		return fmt.Errorf("failed to create server logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to create server config: %w", err)
	}

	repo, err := repository.NewRepository(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	svc := service.NewMetricsService(repo)
	gzip := compress.NewGzip()
	h, err := handler.NewHandler(svc, logger, gzip, cfg)
	if err != nil {
		return fmt.Errorf("failed to create handler: %w", err)
	}

	r := h.NewMetricsRouter()

	server := &http.Server{
		Addr:         cfg.ServerBaseURL,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	appCtx, appCancel := context.WithCancel(ctx)
	defer appCancel()

	go func() {
		logger.Info("server started", zap.String("url", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen error", zap.Error(err))
			appCancel()
		}
	}()

	<-appCtx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("server stopped gracefully")

	return nil
}
