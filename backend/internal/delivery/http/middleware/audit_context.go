package middleware

import (
	"context"
	"net/http"
)

const (
	// AuditIPKey is the context key for the client's IP address.
	AuditIPKey contextKey = "audit_ip"
	// AuditUserAgentKey is the context key for the client's user agent.
	AuditUserAgentKey contextKey = "audit_user_agent"
	// AuditEndpointKey is the context key for the request endpoint path.
	AuditEndpointKey contextKey = "audit_endpoint"
	// AuditMethodKey is the context key for the HTTP method.
	AuditMethodKey contextKey = "audit_method"
)

// AuditContext extracts request metadata into context for audit logging.
func AuditContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, AuditIPKey, r.RemoteAddr)
		ctx = context.WithValue(ctx, AuditUserAgentKey, r.Header.Get("User-Agent"))
		ctx = context.WithValue(ctx, AuditEndpointKey, r.URL.Path)
		ctx = context.WithValue(ctx, AuditMethodKey, r.Method)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
