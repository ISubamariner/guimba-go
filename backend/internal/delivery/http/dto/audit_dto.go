package dto

import (
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type AuditEntryResponse struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	UserEmail    string         `json:"user_email"`
	UserRole     string         `json:"user_role"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	IPAddress    string         `json:"ip_address"`
	Endpoint     string         `json:"endpoint"`
	Method       string         `json:"method"`
	StatusCode   int            `json:"status_code"`
	Success      bool           `json:"success"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Timestamp    string         `json:"timestamp"`
}

type AuditListResponse struct {
	Data   []AuditEntryResponse `json:"data"`
	Total  int                  `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

func NewAuditEntryResponse(e *repository.AuditEntry) AuditEntryResponse {
	return AuditEntryResponse{
		ID:           e.ID.String(),
		UserID:       e.UserID.String(),
		UserEmail:    e.UserEmail,
		UserRole:     e.UserRole,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID.String(),
		IPAddress:    e.IPAddress,
		Endpoint:     e.Endpoint,
		Method:       e.Method,
		StatusCode:   e.StatusCode,
		Success:      e.Success,
		ErrorMessage: e.ErrorMessage,
		Metadata:     e.Metadata,
		Timestamp:    e.Timestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func NewAuditListResponse(entries []*repository.AuditEntry, total, limit, offset int) AuditListResponse {
	data := make([]AuditEntryResponse, 0, len(entries))
	for _, e := range entries {
		data = append(data, NewAuditEntryResponse(e))
	}
	return AuditListResponse{Data: data, Total: total, Limit: limit, Offset: offset}
}
