package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"metrics-collector/internal/app"
	"metrics-collector/internal/config"
	"metrics-collector/internal/logger"
	"metrics-collector/internal/pprof"
	"metrics-collector/pkg/buildinfo"

	"go.uber.org/zap"
)

func main() {
	buildinfo.Print()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("failed to sync logger: %v", err)
		}
	}()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		logger.Fatal("failed to create config: %v", zap.Error(err))
	}

	if cfg == nil {
		logger.Fatal("config is nil")
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

	if err := app.Run(ctx); err != nil {
		logger.Fatal("server app failed: %v", zap.Error(err))
	}
}
