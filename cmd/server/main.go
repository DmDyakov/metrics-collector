package main

import (
	"log"
	"net/http"
	"os"

	"metrics-collector/internal/compress"
	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/logger"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"

	"go.uber.org/zap"
)

func main() {
	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("Failed to create server logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		logger.Fatal("Failed to create server config", zap.Error(err))
	}

	repo, err := repository.NewRepository(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create repository", zap.Error(err))
	}

	svc := service.NewMetricsService(repo)
	gzip := compress.NewGzip()
	h, err := handler.NewHandler(svc, logger, gzip, cfg)
	if err != nil {
		logger.Fatal("Failed to create handler", zap.Error(err))
	}

	r := h.NewMetricsRouter()

	logger.Info("Server started",
		zap.String("url", cfg.ServerBaseURL),
	)

	if err := http.ListenAndServe(cfg.ServerBaseURL, r); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}
