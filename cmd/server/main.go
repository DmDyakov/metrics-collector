package main

import (
	"log"
	"net/http"
	"os"

	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
)

func main() {
	cfg, err := config.NewServerConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to create server config: %v", err)
	}

	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	h := handler.NewHandler(svc)

	r := h.NewMetricsRouter()

	log.Printf("Server started on %s...", cfg.ServerBaseURL)
	log.Fatal(http.ListenAndServe(cfg.ServerBaseURL, r))
}
