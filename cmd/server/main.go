package main

import (
	"log"
	"net/http"

	"metrics-collector/internal/handler"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
)

func main() {
	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", h.UpdateMetrics)

	log.Println("Server started on :8080")
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
