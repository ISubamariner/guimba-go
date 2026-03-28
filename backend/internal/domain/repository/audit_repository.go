package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditEntry represents an immutable audit log record.
type AuditEntry struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	UserEmail    string
	UserRole     string
	Action       string
	ResourceType string
	ResourceID   uuid.UUID
	IPAddress    string
	UserAgent    string
	Endpoint     string
	Method       string
	StatusCode   int
	Success      bool
	ErrorMessage *string
	Metadata     map[string]any
	Timestamp    time.Time
}

// AuditFilter specifies criteria for querying audit logs.
type AuditFilter struct {
	UserID       *uuid.UUID
	LandlordID   *uuid.UUID // user_id OR metadata.landlord_id match
	Action       *string
	ResourceType *string
	Success      *bool
	FromDate     *time.Time
	ToDate       *time.Time
	Limit        int
	Offset       int
}

// AuditRepository defines the interface for audit log persistence.
type AuditRepository interface {
	Log(ctx context.Context, entry *AuditEntry) error
	List(ctx context.Context, filter AuditFilter) ([]*AuditEntry, int, error)
}
