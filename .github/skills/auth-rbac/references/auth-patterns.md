# Auth Patterns Reference

## JWT Claims Structure

```go
type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    Roles  []string  `json:"roles"`
    jwt.RegisteredClaims
}
```

**Registered claims used**:
- `ExpiresAt`: token expiry
- `IssuedAt`: token creation time
- `Issuer`: `"guimba-api"`

## Middleware Chain Example

```go
r.Route("/api/v1", func(r chi.Router) {
    // Public
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/register", authHandler.Register)
    r.Post("/auth/refresh", authHandler.RefreshToken)

    // Authenticated
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthMiddleware)

        // Any authenticated user
        r.Get("/me", userHandler.GetProfile)
        r.Put("/me", userHandler.UpdateProfile)

        // Staff and above
        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireRole("staff", "admin"))
            r.Post("/programs", programHandler.Create)
            r.Put("/programs/{id}", programHandler.Update)
            r.Delete("/programs/{id}", programHandler.Delete)
            r.Post("/beneficiaries", beneficiaryHandler.Create)
            r.Put("/beneficiaries/{id}", beneficiaryHandler.Update)
            r.Delete("/beneficiaries/{id}", beneficiaryHandler.Delete)
            r.Post("/beneficiaries/{id}/programs", beneficiaryHandler.EnrollInProgram)
            r.Delete("/beneficiaries/{id}/programs/{programId}", beneficiaryHandler.RemoveFromProgram)
        })

        // Manager and above — reserved for future debt/beneficiary approval
        // (Currently no "manager" role; only admin/staff/viewer are seeded)

        // Admin only
        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireRole("admin"))
            r.Get("/users", userHandler.List)
            r.Put("/users/{id}", userHandler.Update)
            r.Delete("/users/{id}", userHandler.Delete)
            r.Post("/users/{id}/roles", userHandler.AssignRole)
        })
    })
})
```

## Role Hierarchy

```
admin  → full system access, user management, role assignment
staff  → create/edit programs and beneficiaries, read users
viewer → read-only access to dashboards and reports
```

3 system roles are seeded via migration `000004_seed_system_roles.up.sql` with 13 permissions.
Roles are stored in a junction table (`user_roles`) and eagerly loaded with the User entity.

## Refresh Token Rotation

```go
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    // 1. Extract refresh token from httpOnly cookie
    cookie, err := r.Cookie("refresh_token")
    if err != nil {
        respondError(w, apperror.NewUnauthorized("missing refresh token"))
        return
    }

    // 2. Validate refresh token
    claims, err := auth.ValidateRefreshToken(cookie.Value)
    if err != nil {
        respondError(w, apperror.NewUnauthorized("invalid refresh token"))
        return
    }

    // 3. Load user (ensures user still exists and is active)
    user, err := h.userUseCase.GetByID(r.Context(), claims.UserID)
    if err != nil {
        respondError(w, apperror.NewUnauthorized("user not found"))
        return
    }

    // 4. Generate new token pair (rotation)
    accessToken, refreshToken, err := auth.GenerateTokenPair(user)
    if err != nil {
        respondError(w, apperror.NewInternal("token generation failed"))
        return
    }

    // 5. Set new refresh token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        Path:     "/api/v1/auth",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   7 * 24 * 60 * 60, // 7 days
    })

    // 6. Return new access token
    respondJSON(w, http.StatusOK, map[string]string{
        "access_token": accessToken,
    })
}
```

## Protected Route Registration Example

```go
// In router/router.go
func NewRouter(
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    debtHandler *handler.DebtHandler,
    authMiddleware func(http.Handler) http.Handler,
) chi.Router {
    r := chi.NewRouter()

    // Global middleware
    r.Use(chimiddleware.Logger)
    r.Use(chimiddleware.Recoverer)
    r.Use(cors.Handler(corsOptions))

    r.Route("/api/v1", func(r chi.Router) {
        // Health check (public)
        r.Get("/health", handler.HealthCheck)

        // Auth routes (public)
        r.Route("/auth", func(r chi.Router) {
            r.Post("/login", authHandler.Login)
            r.Post("/register", authHandler.Register)
            r.Post("/refresh", authHandler.RefreshToken)
        })

        // All other routes require auth
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware)

            r.Route("/users", func(r chi.Router) {
                r.Use(middleware.RequireRole("admin"))
                r.Get("/", userHandler.List)
                r.Get("/{id}", userHandler.GetByID)
                r.Delete("/{id}", userHandler.Delete)
            })

            r.Route("/debts", func(r chi.Router) {
                r.Get("/", debtHandler.List)
                r.Post("/", debtHandler.Create)
                r.Get("/{id}", debtHandler.GetByID)
                r.Put("/{id}", debtHandler.Update)
            })
        })
    })

    return r
}
```
