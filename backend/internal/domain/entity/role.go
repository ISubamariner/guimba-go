package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role represents an RBAC role.
type Role struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	DisplayName  string       `json:"display_name"`
	Description  string       `json:"description"`
	IsSystemRole bool         `json:"is_system_role"`
	IsActive     bool         `json:"is_active"`
	Permissions  []Permission `json:"permissions,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Permission represents a granular permission.
type Permission struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	DisplayName        string    `json:"display_name"`
	Description        string    `json:"description"`
	Category           string    `json:"category"`
	IsSystemPermission bool      `json:"is_system_permission"`
	CreatedAt          time.Time `json:"created_at"`
}

// Validate checks business rules for a Role.
func (r *Role) Validate() error {
	if r.Name == "" {
		return ErrRoleNameRequired
	}
	if len(r.Name) > 50 {
		return ErrRoleNameTooLong
	}
	if r.DisplayName == "" {
		return ErrRoleDisplayNameRequired
	}
	return nil
}
