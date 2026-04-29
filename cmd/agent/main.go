package main

import (
	"context"
	"log"
	"metrics-collector/internal/agent"
	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"
	"metrics-collector/internal/logger"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("Failed to create agent logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.NewAgentConfig(os.Args[1:])
	if err != nil {
		logger.Fatal("failed to create agent config", zap.Error(err))
	}

	if cfg == nil {
		logger.Fatal("Config is nil")
	}

	gzip := compress.NewGzip()
	agent := agent.NewAgent(cfg, logger, gzip)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	agent.Run(ctx)
}
