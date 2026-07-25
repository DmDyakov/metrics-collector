package app

import (
	"metrics-collector/internal/audit"
	"metrics-collector/internal/config"
	"metrics-collector/internal/handler"
	"metrics-collector/internal/middleware"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// registerRoutes registers all HTTP server routes with middleware.
func registerRoutes(
	healthHandler *handler.HealthHandler,
	metricsHandler *handler.MetricsHandler,
	auditPublisher *audit.Publisher,
	signer *signer.Signer,
	encryptor *encryptor.Encryptor,
	logger *zap.Logger,
	cfg *config.ServerConfig,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.WithTimeout(cfg.RequestTimeout))
	r.Use(middleware.WithLogging(logger))

	r.Use(middleware.WithSignature(logger, signer))
	r.Use(middleware.WithDecryption(logger, encryptor))
	r.Use(middleware.WithCompressing)

	r.Get("/ping", healthHandler.HealthDB)

	r.Group(func(r chi.Router) {
		r.Get("/", metricsHandler.ListMetrics)
		r.Get("/value/{type}/{name}", metricsHandler.GetMetricValue)
		r.Post("/value", metricsHandler.GetMetric)
	})

	r.Group(func(r chi.Router) {
		if auditPublisher != nil {
			r.Use(middleware.WithAudit(logger, auditPublisher))
		}

		r.Post("/update/{type}/{name}/{value}", metricsHandler.UpdateMetricByURL)
		r.Post("/update", metricsHandler.UpdateMetricByJSON)
		r.Post("/updates", metricsHandler.UpdateMetricsBatch)
	})

	return r
}
