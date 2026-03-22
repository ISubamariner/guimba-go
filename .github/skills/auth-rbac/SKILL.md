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
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    Roles  []string  `json:"roles"`
    jwt.RegisteredClaims
}

// JWTManager handles token generation and validation
func NewJWTManager(secret string, accessDuration, refreshDuration time.Duration) *JWTManager
func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, email string, roles []string) (*TokenPair, error)
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error)
```

### Token Blocklist (`infrastructure/cache/token_blocklist.go`)
On logout or token refresh, old tokens are blocklisted in Redis with a TTL matching the token's remaining lifetime:
```go
blocklist.Block(ctx, claims.ID, remainingTTL)   // Add to blocklist
blocklist.IsBlocked(ctx, claims.ID) (bool, error) // Check before processing
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
// AuthMiddleware validates JWT from Authorization header and checks blocklist
func AuthMiddleware(jwtManager *auth.JWTManager, blocklist *cache.TokenBlocklist) func(http.Handler) http.Handler

// RequireRole checks that the authenticated user has at least one of the specified roles
func RequireRole(roles ...string) func(http.Handler) http.Handler
```

Context keys set by auth middleware:
- `AuthUserIDKey` → `uuid.UUID`
- `AuthEmailKey` → `string`
- `AuthRolesKey` → `[]string`

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
admin  → full system access, user management, role assignment
staff  → create/edit programs and beneficiaries, view users (read-only)
viewer → read-only access to dashboards and reports
```

> **Note**: 3 system roles are seeded via migration `000004_seed_system_roles.up.sql` with 13 permissions across programs, users, and beneficiaries categories.

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
