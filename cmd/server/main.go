package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"metrics-collector/internal/config"
	"metrics-collector/internal/server/app"
	"metrics-collector/pkg/buildinfo"
	"metrics-collector/pkg/lifecycle"
	"metrics-collector/pkg/logger"
	"metrics-collector/pkg/pprof"

	"go.uber.org/zap"
)

func main() {
	buildinfo.Print()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
		}
	}()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		logger.Fatal("failed to create config: %v", zap.Error(err))
	}

	app, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create app: %v", zap.Error(err))
	}

	if cfg.PprofAddr != "" {
		go func() {
			logger.Info("pprof server started", zap.String("addr", cfg.PprofAddr))
			if err := pprof.Serve(cfg.PprofAddr); err != nil {
				logger.Error("pprof server failed", zap.Error(err))
			}
		}()
	}

	if err := lifecycle.Run(ctx, app, cfg.ShutdownTimeout); err != nil {
		logger.Fatal("app terminated with error", zap.Error(err))
	}

	logger.Info("App stopped gracefully")
}
