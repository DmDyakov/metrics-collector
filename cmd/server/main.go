package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"metrics-collector/internal/app"
	"metrics-collector/internal/config"
	"metrics-collector/internal/logger"
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
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	app, err := app.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	return app.Run(ctx)
}
