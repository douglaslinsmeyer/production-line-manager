package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"ping/production-line-api/internal/service"
)

// NewRouter creates and configures the HTTP router
func NewRouter(lineService *service.LineService, corsOrigins string, logger *zap.Logger) *chi.Mux {
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
			})
		})
	})

	return r
}
