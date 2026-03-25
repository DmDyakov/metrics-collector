package main

import (
	"log"
	"net/http"
	"os"

	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"

	"metrics-collector/internal/logger"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
)

func main() {
	logger, err := logger.NewSugarZapLogger()
	if err != nil {
		log.Fatalf("Failed to create server logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		logger.Fatalf("Failed to create server config: %v", err)
	}

	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	h, err := handler.NewHandler(svc, logger)
	if err != nil {
		logger.Fatalw("server failed", "error", err)
	}

	r := h.NewMetricsRouter()

	logger.Infof("Server started on %s...", cfg.ServerBaseURL)

	if err := http.ListenAndServe(cfg.ServerBaseURL, r); err != nil {
		logger.Fatalw("server failed", "error", err)
	}
}
