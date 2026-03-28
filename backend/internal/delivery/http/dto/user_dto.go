package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	FullName string `json:"full_name" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest is the request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse is the response body after login or registration.
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// TokenResponse is the response for a token refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserResponse is the response body for a single user.
type UserResponse struct {
	ID              uuid.UUID      `json:"id"`
	Email           string         `json:"email"`
	FullName        string         `json:"full_name"`
	IsActive        bool           `json:"is_active"`
	IsEmailVerified bool           `json:"is_email_verified"`
	LastLoginAt     *string        `json:"last_login_at,omitempty"`
	Roles           []RoleResponse `json:"roles"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	FullName string `json:"full_name" validate:"required,max=255"`
	IsActive bool   `json:"is_active"`
}

// AssignRoleRequest is the request body for assigning/removing a role.
type AssignRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// ChangePasswordRequest is the request body for changing user password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
}

// UserListResponse is the response body for a list of users.
type UserListResponse struct {
	Data   []UserResponse `json:"data"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// RoleResponse is the response body for a role.
type RoleResponse struct {
	ID           uuid.UUID            `json:"id"`
	Name         string               `json:"name"`
	DisplayName  string               `json:"display_name"`
	Description  string               `json:"description"`
	IsSystemRole bool                 `json:"is_system_role"`
	Permissions  []PermissionResponse `json:"permissions,omitempty"`
}

// PermissionResponse is the response body for a permission.
type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Category    string    `json:"category"`
}

// NewUserResponse creates a UserResponse from a domain entity.
func NewUserResponse(u *entity.User) UserResponse {
	resp := UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		FullName:        u.FullName,
		IsActive:        u.IsActive,
		IsEmailVerified: u.IsEmailVerified,
		Roles:           make([]RoleResponse, 0, len(u.Roles)),
		CreatedAt:       u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       u.UpdatedAt.Format(time.RFC3339),
	}
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &s
	}
	for _, r := range u.Roles {
		resp.Roles = append(resp.Roles, NewRoleResponse(&r))
	}
	return resp
}

// NewRoleResponse creates a RoleResponse from a domain entity.
func NewRoleResponse(r *entity.Role) RoleResponse {
	resp := RoleResponse{
		ID:           r.ID,
		Name:         r.Name,
		DisplayName:  r.DisplayName,
		Description:  r.Description,
		IsSystemRole: r.IsSystemRole,
		Permissions:  make([]PermissionResponse, 0, len(r.Permissions)),
	}
	for _, p := range r.Permissions {
		resp.Permissions = append(resp.Permissions, PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Category:    p.Category,
		})
	}
	return resp
}

// NewUserListResponse creates a UserListResponse from domain entities.
func NewUserListResponse(users []*entity.User, total, limit, offset int) UserListResponse {
	data := make([]UserResponse, 0, len(users))
	for _, u := range users {
		data = append(data, NewUserResponse(u))
	}
	return UserListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
