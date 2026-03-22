package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

const (
	// AuthUserIDKey is the context key for the authenticated user's ID.
	AuthUserIDKey contextKey = "auth_user_id"
	// AuthEmailKey is the context key for the authenticated user's email.
	AuthEmailKey contextKey = "auth_email"
	// AuthRolesKey is the context key for the authenticated user's roles.
	AuthRolesKey contextKey = "auth_roles"
)

// AuthMiddleware validates the JWT access token from the Authorization header.
func AuthMiddleware(jwtManager *auth.JWTManager, blocklist *cache.TokenBlocklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				apperror.WriteError(w, apperror.NewUnauthorized("Missing authorization header"))
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				apperror.WriteError(w, apperror.NewUnauthorized("Invalid authorization header format"))
				return
			}

			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				apperror.WriteError(w, apperror.NewUnauthorized("Invalid or expired token"))
				return
			}

			// Check token blocklist
			if blocklist != nil {
				blocked, err := blocklist.IsBlocked(r.Context(), claims.ID)
				if err == nil && blocked {
					apperror.WriteError(w, apperror.NewUnauthorized("Token has been revoked"))
					return
				}
			}

			// Set auth context values
			ctx := context.WithValue(r.Context(), AuthUserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, AuthEmailKey, claims.Email)
			ctx = context.WithValue(ctx, AuthRolesKey, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that ensures the authenticated user has at least one of the specified roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := r.Context().Value(AuthRolesKey).([]string)
			if !ok || len(userRoles) == 0 {
				apperror.WriteError(w, apperror.NewForbidden("Insufficient permissions"))
				return
			}

			for _, required := range roles {
				for _, has := range userRoles {
					if has == required {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			apperror.WriteError(w, apperror.NewForbidden("Insufficient permissions"))
		})
	}
}
