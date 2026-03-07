package main

import (
	"fmt"
	"log"
	"net/http"

	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/repository"
	"metrics-collector/internal/service"
)

func main() {
	cfg := config.NewConfig()

	repo := repository.NewMemStorage()
	svc := service.NewMetricsService(repo)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", h.UpdateMetrics)

	log.Printf("Server started on :%s...", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), logRequest(mux)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}
