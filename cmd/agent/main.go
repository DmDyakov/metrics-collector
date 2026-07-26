package main

import (
	"context"
	"fmt"
	"log"
	"metrics-collector/internal/agent"
	"metrics-collector/internal/config"
	"metrics-collector/pkg/buildinfo"
	"metrics-collector/pkg/logger"
	"os"
	"os/signal"
	"syscall"

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

	cfg, err := config.NewAgentConfig(os.Args[1:])
	if err != nil {
		logger.Fatal("failed to create agent config", zap.Error(err))
	}

	if cfg == nil {
		logger.Fatal("Config is nil")
	}

	agent := agent.NewAgent(cfg, logger)

	if err := agent.Run(ctx); err != nil {
		logger.Fatal("agent stopped with error", zap.Error(err))
	}

	logger.Info("Agent stopped gracefully")
}
