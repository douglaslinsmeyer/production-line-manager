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
	scheduleService *service.ScheduleService,
	complianceService *service.ComplianceService,
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
	scheduleHandler := NewScheduleHandler(scheduleService, logger)
	complianceHandler := NewComplianceHandler(complianceService, logger)

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

				// Schedule assignment
				r.Put("/schedule", scheduleHandler.AssignScheduleToLine)
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
				r.Get("/compliance/daily", complianceHandler.GetDailyComplianceKPIs)
			})

			// Schedule compliance endpoints
			r.Get("/compliance", complianceHandler.GetAggregateCompliance)
			r.Get("/compliance/lines", complianceHandler.GetLineCompliance)
		})

		// Schedules
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", scheduleHandler.List)
			r.Post("/", scheduleHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", scheduleHandler.Get)
				r.Put("/", scheduleHandler.Update)
				r.Delete("/", scheduleHandler.Delete)

				// Schedule days
				r.Route("/days/{dayId}", func(r chi.Router) {
					r.Get("/", scheduleHandler.GetDay)
					r.Put("/", scheduleHandler.UpdateDay)

					// Day breaks
					r.Put("/breaks", scheduleHandler.SetDayBreaks)
				})

				// Holidays
				r.Route("/holidays", func(r chi.Router) {
					r.Get("/", scheduleHandler.ListHolidays)
					r.Post("/", scheduleHandler.CreateHoliday)

					r.Route("/{holidayId}", func(r chi.Router) {
						r.Get("/", scheduleHandler.GetHoliday)
						r.Put("/", scheduleHandler.UpdateHoliday)
						r.Delete("/", scheduleHandler.DeleteHoliday)
					})
				})

				// Schedule exceptions (apply to all lines)
				r.Route("/exceptions", func(r chi.Router) {
					r.Get("/", scheduleHandler.ListExceptions)
					r.Post("/", scheduleHandler.CreateException)

					r.Route("/{exceptionId}", func(r chi.Router) {
						r.Get("/", scheduleHandler.GetException)
						r.Put("/", scheduleHandler.UpdateException)
						r.Delete("/", scheduleHandler.DeleteException)
					})
				})

				// Line-specific exceptions
				r.Route("/line-exceptions", func(r chi.Router) {
					r.Get("/", scheduleHandler.ListLineExceptions)
					r.Post("/", scheduleHandler.CreateLineException)

					r.Route("/{exceptionId}", func(r chi.Router) {
						r.Get("/", scheduleHandler.GetLineException)
						r.Put("/", scheduleHandler.UpdateLineException)
						r.Delete("/", scheduleHandler.DeleteLineException)
					})
				})

				// Lines using this schedule
				r.Get("/lines", scheduleHandler.GetLinesForSchedule)
			})
		})

		// Effective schedule (computed)
		r.Get("/effective-schedule", scheduleHandler.GetEffectiveSchedule)
		r.Get("/effective-schedule/range", scheduleHandler.GetEffectiveScheduleRange)
	})

	return r
}
