package router

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/ISubamariner/guimba-go/backend/docs"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
)

// NewRouter creates and configures the Chi router with all middleware and routes.
func NewRouter(healthHandler *handler.HealthHandler, frontendURL string) chi.Router {
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (outside /api/v1 — always accessible)
	r.Get("/health", healthHandler.Health)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Future routes will be registered here:
		// r.Route("/programs", func(r chi.Router) { ... })
		// r.Route("/users", func(r chi.Router) { ... })
	})

	return r
}
