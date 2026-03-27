package entity

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a person who owes money to a landlord.
type Tenant struct {
	ID          uuid.UUID  `json:"id"`
	FullName    string     `json:"full_name"`
	Email       *string    `json:"email,omitempty"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	NationalID  *string    `json:"national_id,omitempty"`
	Address     *Address   `json:"address,omitempty"`
	LandlordID  uuid.UUID  `json:"landlord_id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewTenant creates a new Tenant with generated ID and defaults.
func NewTenant(fullName string, email, phoneNumber, nationalID *string, address *Address, landlordID uuid.UUID, notes *string) (*Tenant, error) {
	t := &Tenant{
		ID:          uuid.New(),
		FullName:    fullName,
		Email:       email,
		PhoneNumber: phoneNumber,
		NationalID:  nationalID,
		Address:     address,
		LandlordID:  landlordID,
		IsActive:    true,
		Notes:       notes,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t, nil
}

// Validate checks business rules for a Tenant.
func (t *Tenant) Validate() error {
	if t.FullName == "" {
		return ErrTenantFullNameRequired
	}
	if len(t.FullName) > 255 {
		return ErrTenantFullNameTooLong
	}
	hasEmail := t.Email != nil && *t.Email != ""
	hasPhone := t.PhoneNumber != nil && *t.PhoneNumber != ""
	if !hasEmail && !hasPhone {
		return ErrTenantContactRequired
	}
	return nil
}
