---
name: auth-rbac
description: "Manages authentication (JWT), authorization (RBAC), login/register flows, password hashing, middleware guards, and token refresh. Use when user says 'add auth', 'create login', 'protect route', 'add role', 'JWT', 'middleware guard', 'password hash', 'token refresh', 'sign up', 'register', or when working with middleware/auth*, usecase/user/, or any protected endpoint."
---

# Authentication & Role-Based Access Control

Manages JWT authentication, RBAC authorization, and all auth-related flows.

## JWT Token Flow

### Token Pair Strategy
- **Access token**: short-lived (15 min), used for API requests
- **Refresh token**: long-lived (7 days), used to obtain new access tokens
- Access token stored in memory (frontend) or httpOnly cookie
- Refresh token stored in httpOnly, Secure, SameSite=Strict cookie

### Token Generation (`pkg/auth/jwt.go`)
```go
type Claims struct {
    UserID string   `json:"user_id"`
    Email  string   `json:"email"`
    Roles  []string `json:"roles"`
    jwt.RegisteredClaims
}

func GenerateTokenPair(user *domain.User) (accessToken, refreshToken string, err error)
func ValidateAccessToken(tokenString string) (*Claims, error)
func ValidateRefreshToken(tokenString string) (*Claims, error)
```

## Password Hashing

Always use `golang.org/x/crypto/bcrypt`:
```go
func HashPassword(password string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
```

**Rules**:
- Never store plaintext passwords
- Never log passwords or tokens
- Use `bcrypt.DefaultCost` (10) minimum
- Always validate password strength at the DTO layer

## Auth Middleware (`internal/delivery/http/middleware/auth.go`)

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractBearerToken(r)
        if token == "" {
            respondError(w, apperror.NewUnauthorized("missing auth token"))
            return
        }
        claims, err := auth.ValidateAccessToken(token)
        if err != nil {
            respondError(w, apperror.NewUnauthorized("invalid or expired token"))
            return
        }
        ctx := context.WithValue(r.Context(), userClaimsKey, claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Role-Based Access Control

### Role Check Middleware
```go
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := GetClaims(r.Context())
            if !hasAnyRole(claims.Roles, roles) {
                respondError(w, apperror.NewForbidden("insufficient permissions"))
                return
            }
            next.ServeHTTP(w, r.WithContext(r.Context()))
        })
    }
}
```

### Role Hierarchy
```
admin > manager > staff > viewer
```

### Route Registration with Auth
```go
r.Route("/api/v1", func(r chi.Router) {
    // Public routes
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/register", authHandler.Register)
    r.Post("/auth/refresh", authHandler.RefreshToken)

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthMiddleware)
        r.Get("/profile", userHandler.GetProfile)

        // Admin-only routes
        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireRole("admin"))
            r.Delete("/users/{id}", userHandler.DeleteUser)
        })
    })
})
```

## Login/Register Flow (Clean Architecture)

```
POST /api/v1/auth/login
  → handler.Login() parses LoginDTO
  → usecase.Login(email, password)
  → repo.FindByEmail(email) → domain.User
  → auth.CheckPassword(user.HashedPassword, password)
  → auth.GenerateTokenPair(user)
  → return tokens + user DTO
```

## Frontend Auth

- Store access token in memory (React state/context)
- Refresh token in httpOnly cookie (set by backend)
- Inject `Authorization: Bearer <token>` header on every API call via `src/lib/api.ts`
- On 401 response: attempt token refresh, if fails → redirect to `/login`
- Clear auth state on logout

## Security Rules

- Never log tokens or passwords
- Never store plaintext passwords in DB or seed files (use pre-hashed bcrypt)
- Always validate token expiry
- Use HTTPS in production
- Set Secure, HttpOnly, SameSite=Strict on auth cookies
- Implement rate limiting on auth endpoints
- Invalidate refresh tokens on password change
