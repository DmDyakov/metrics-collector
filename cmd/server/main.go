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

	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		logger.Fatal("failed to create config: %v", zap.Error(err))
	}

	if cfg == nil {
		logger.Fatal("Config is nil")
	}

	app, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create app: %v", zap.Error(err))
	}

	if err := app.Run(ctx); err != nil {
		logger.Fatal("server app failed: %v", zap.Error(err))
	}
}
