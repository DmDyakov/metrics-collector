package handler

import (
	"html/template"

	"metrics-collector/internal/compress"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/service"
	"metrics-collector/internal/templates"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricsService interface {
	UpdateMetric(metricType, metricName, metricValue string) error
	UpdateMetricV2(metric models.Metrics) (*models.Metrics, error)
	GetMetricValue(metricType, metricName string) (string, error)
	GetMetricValueV2(m models.Metrics) (*models.Metrics, error)
	GetAllMetrics() (map[string]string, error)
}

type Handler struct {
	service                *service.MetricsService
	logger                 *zap.SugaredLogger
	gzip                   *compress.Gzip
	allMetricsHTMLTemplate *template.Template
}

func NewHandler(service *service.MetricsService, logger *zap.SugaredLogger, gzip *compress.Gzip) (*Handler, error) {
	tmpl, err := template.ParseFS(templates.FS, "metrics.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		service:                service,
		logger:                 logger,
		gzip:                   gzip,
		allMetricsHTMLTemplate: tmpl,
	}, nil
}

func (h *Handler) NewMetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(h.WithLogging)
	r.Use(h.WithCompressing)

	r.Get("/", h.RootHandle)
	r.Get("/value/{type}/{name}", h.ValueHandle)
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandle)
	r.Post("/value", h.ValueHandleV2)
	r.Post("/update", h.UpdateHandleV2)

	return r
}
