package audit

import (
	"context"
	"strings"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/google/uuid"
)

// RequestInfo holds audit-relevant metadata extracted from context.
type RequestInfo struct {
	UserID    uuid.UUID
	UserEmail string
	UserRole  string
	IPAddress string
	UserAgent string
	Endpoint  string
	Method    string
}

// FromContext extracts audit-relevant fields from context.
// Returns safe defaults when keys are missing (e.g., in unit tests).
func FromContext(ctx context.Context) RequestInfo {
	info := RequestInfo{}

	if v, ok := ctx.Value(middleware.AuthUserIDKey).(string); ok {
		if parsed, err := uuid.Parse(v); err == nil {
			info.UserID = parsed
		}
	}
	if v, ok := ctx.Value(middleware.AuthEmailKey).(string); ok {
		info.UserEmail = v
	}
	if v, ok := ctx.Value(middleware.AuthRolesKey).([]string); ok {
		info.UserRole = strings.Join(v, ",")
	}
	if v, ok := ctx.Value(middleware.AuditIPKey).(string); ok {
		info.IPAddress = v
	}
	if v, ok := ctx.Value(middleware.AuditUserAgentKey).(string); ok {
		info.UserAgent = v
	}
	if v, ok := ctx.Value(middleware.AuditEndpointKey).(string); ok {
		info.Endpoint = v
	}
	if v, ok := ctx.Value(middleware.AuditMethodKey).(string); ok {
		info.Method = v
	}

	return info
}
