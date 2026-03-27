package router

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/ISubamariner/guimba-go/backend/docs"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

// Handlers holds all HTTP handlers for route registration.
type Handlers struct {
	Health      *handler.HealthHandler
	Program     *handler.ProgramHandler
	Auth        *handler.AuthHandler
	User        *handler.UserHandler
	Beneficiary *handler.BeneficiaryHandler
	Tenant      *handler.TenantHandler
}

// NewRouter creates and configures the Chi router with all middleware and routes.
func NewRouter(h Handlers, frontendURL string, jwtManager *auth.JWTManager, blocklist *cache.TokenBlocklist) chi.Router {
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
	r.Get("/health", h.Health.Health)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Auth middleware
	requireAuth := middleware.AuthMiddleware(jwtManager, blocklist)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Auth.Register)
			r.Post("/login", h.Auth.Login)
			r.Post("/refresh", h.Auth.Refresh)

			// Authenticated auth routes
			r.Group(func(r chi.Router) {
				r.Use(requireAuth)
				r.Get("/me", h.Auth.Me)
				r.Post("/logout", h.Auth.Logout)
			})
		})

		// Programs (public read, authenticated write)
		r.Route("/programs", func(r chi.Router) {
			r.Get("/", h.Program.List)
			r.Get("/{id}", h.Program.Get)

			r.Group(func(r chi.Router) {
				r.Use(requireAuth)
				r.Use(middleware.RequireRole("admin", "staff"))
				r.Post("/", h.Program.Create)
				r.Put("/{id}", h.Program.Update)
				r.Delete("/{id}", h.Program.Delete)
			})
		})

		// Users (admin only)
		r.Route("/users", func(r chi.Router) {
			r.Use(requireAuth)
			r.Use(middleware.RequireRole("admin"))
			r.Get("/", h.User.List)
			r.Put("/{id}", h.User.Update)
			r.Delete("/{id}", h.User.Delete)
			r.Post("/{id}/roles", h.User.AssignRole)
			r.Delete("/{id}/roles/{roleId}", h.User.RemoveRole)
		})

		// Beneficiaries (authenticated read, staff+ write)
		r.Route("/beneficiaries", func(r chi.Router) {
			r.Use(requireAuth)
			r.Get("/", h.Beneficiary.List)
			r.Get("/{id}", h.Beneficiary.Get)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "staff"))
				r.Post("/", h.Beneficiary.Create)
				r.Put("/{id}", h.Beneficiary.Update)
				r.Delete("/{id}", h.Beneficiary.Delete)
				r.Post("/{id}/programs", h.Beneficiary.EnrollInProgram)
				r.Delete("/{id}/programs/{programId}", h.Beneficiary.RemoveFromProgram)
			})
		})

		// Tenants (authenticated, landlord/admin)
		r.Route("/tenants", func(r chi.Router) {
			r.Use(requireAuth)
			r.Use(middleware.RequireRole("admin", "landlord"))
			r.Get("/", h.Tenant.List)
			r.Get("/{id}", h.Tenant.Get)
			r.Post("/", h.Tenant.Create)
			r.Put("/{id}", h.Tenant.Update)
			r.Put("/{id}/deactivate", h.Tenant.Deactivate)
			r.Delete("/{id}", h.Tenant.Delete)
		})
	})

	return r
}
