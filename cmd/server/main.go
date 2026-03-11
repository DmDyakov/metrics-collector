package main

import (
	"log"
	"net/http"

	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
)

func main() {
	cfg := config.NewServerConfig()

	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	h := handler.NewHandler(svc)

	r := h.NewMetricsRouter()

	log.Printf("Server started on :%s...", cfg.ServerBaseURL)
	log.Fatal(http.ListenAndServe(cfg.ServerBaseURL, r))
}
