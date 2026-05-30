package main

import (
	"context"
	"log"
	"metrics-collector/internal/agent"
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

	agent := agent.NewAgent(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := agent.Run(ctx); err != nil {
		logger.Fatal("agent stopped with error", zap.Error(err))
	}

	logger.Info("Agent stopped gracefully")
}
