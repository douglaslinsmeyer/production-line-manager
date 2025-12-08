package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"ping/production-line-api/internal/service"
)

// NewRouter creates and configures the HTTP router
func NewRouter(
	lineService *service.LineService,
	labelService *service.LabelService,
	analyticsService *service.AnalyticsService,
	corsOrigins string,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS middleware (if configured)
	if corsOrigins != "" {
		r.Use(CORSMiddleware(corsOrigins))
	}

	// Create handlers
	healthHandler := NewHealthHandler()
	lineHandler := NewLineHandler(lineService, logger)
	labelHandler := NewLabelHandler(labelService, logger)
	analyticsHandler := NewAnalyticsHandler(analyticsService, logger)

	// Health check
	r.Get("/health", healthHandler.Health)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Production lines
		r.Route("/lines", func(r chi.Router) {
			r.Get("/", lineHandler.List)
			r.Post("/", lineHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", lineHandler.Get)
				r.Put("/", lineHandler.Update)
				r.Delete("/", lineHandler.Delete)

				// Status endpoints
				r.Post("/status", lineHandler.SetStatus)
				r.Get("/status/history", lineHandler.GetStatusHistory)

				// Label endpoints
				r.Get("/labels", labelHandler.GetLabelsForLine)
				r.Put("/labels", labelHandler.AssignToLine)
			})
		})

		// Labels
		r.Route("/labels", func(r chi.Router) {
			r.Get("/", labelHandler.List)
			r.Post("/", labelHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", labelHandler.Get)
				r.Put("/", labelHandler.Update)
				r.Delete("/", labelHandler.Delete)
			})
		})

		// Analytics
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/aggregate", analyticsHandler.GetAggregateMetrics)
			r.Get("/lines", analyticsHandler.GetLineMetrics)
			r.Get("/labels", analyticsHandler.GetLabelMetrics)

			r.Route("/lines/{id}", func(r chi.Router) {
				r.Get("/daily", analyticsHandler.GetDailyKPIs)
			})
		})
	})

	return r
}
