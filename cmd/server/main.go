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
	"metrics-collector/internal/worker"
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

	repo, err := repository.NewRepository(cfg)
	if err != nil {
		logger.Fatalw("Failed to create repository", "error", err)
	}

	if cfg.StoreInterval > 0 {
		backupWorker := worker.NewBackupWorker(cfg.StoreInterval, repo, logger)
		backupWorker.Start()
		logger.Infof("Backup worker started with interval %d seconds", cfg.StoreInterval)
	} else {
		logger.Infof("Backup worker disabled (store_interval = 0)")
	}

	svc := service.NewMetricsService(repo)
	gzip := compress.NewGzip()
	h, err := handler.NewHandler(svc, logger, gzip)
	if err != nil {
		logger.Fatalw("Failed to create handler", "error", err)
	}

	r := h.NewMetricsRouter()

	logger.Infof("Server started on %s...", cfg.ServerBaseURL)

	if err := http.ListenAndServe(cfg.ServerBaseURL, r); err != nil {
		logger.Fatalw("server failed", "error", err)
	}
}
