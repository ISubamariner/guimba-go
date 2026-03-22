package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user.
type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	HashedPassword  string     `json:"-"`
	IsActive        bool       `json:"is_active"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	Roles           []Role     `json:"roles,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// NewUser creates a new User with generated ID and timestamps.
func NewUser(email, fullName, hashedPassword string) (*User, error) {
	u := &User{
		ID:              uuid.New(),
		Email:           email,
		FullName:        fullName,
		HashedPassword:  hashedPassword,
		IsActive:        true,
		IsEmailVerified: false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

// Validate checks business rules for a User.
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrUserEmailRequired
	}
	if len(u.Email) > 255 {
		return ErrUserEmailTooLong
	}
	if u.FullName == "" {
		return ErrUserFullNameRequired
	}
	if len(u.FullName) > 255 {
		return ErrUserFullNameTooLong
	}
	if u.HashedPassword == "" {
		return ErrUserPasswordRequired
	}
	return nil
}

// HasRole checks if the user has a specific role by name.
func (u *User) HasRole(roleName string) bool {
	for _, r := range u.Roles {
		if r.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission through any of their roles.
func (u *User) HasPermission(permName string) bool {
	for _, r := range u.Roles {
		for _, p := range r.Permissions {
			if p.Name == permName {
				return true
			}
		}
	}
	return false
}

// HasAnyRole checks if the user has at least one of the given roles.
func (u *User) HasAnyRole(roleNames ...string) bool {
	for _, name := range roleNames {
		if u.HasRole(name) {
			return true
		}
	}
	return false
}
