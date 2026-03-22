package entity

import (
	"time"

	"github.com/google/uuid"
)

// BeneficiaryStatus represents the possible statuses of a beneficiary.
type BeneficiaryStatus string

const (
	BeneficiaryStatusActive    BeneficiaryStatus = "active"
	BeneficiaryStatusInactive  BeneficiaryStatus = "inactive"
	BeneficiaryStatusSuspended BeneficiaryStatus = "suspended"
)

// Beneficiary represents a person who receives benefits from social programs.
type Beneficiary struct {
	ID          uuid.UUID           `json:"id"`
	FullName    string              `json:"full_name"`
	Email       *string             `json:"email,omitempty"`
	PhoneNumber *string             `json:"phone_number,omitempty"`
	NationalID  *string             `json:"national_id,omitempty"`
	Address     *string             `json:"address,omitempty"`
	DateOfBirth *time.Time          `json:"date_of_birth,omitempty"`
	Status      BeneficiaryStatus   `json:"status"`
	Notes       *string             `json:"notes,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   *time.Time          `json:"deleted_at,omitempty"`
	Programs    []ProgramEnrollment `json:"programs,omitempty"`
}

// ProgramEnrollment represents a beneficiary's enrollment in a program.
type ProgramEnrollment struct {
	ProgramID   uuid.UUID `json:"program_id"`
	ProgramName string    `json:"program_name"`
	EnrolledAt  time.Time `json:"enrolled_at"`
	Status      string    `json:"status"`
}

// NewBeneficiary creates a new Beneficiary with generated ID and timestamps.
func NewBeneficiary(fullName string, email, phoneNumber, nationalID, address *string, dateOfBirth *time.Time, status BeneficiaryStatus, notes *string) (*Beneficiary, error) {
	b := &Beneficiary{
		ID:          uuid.New(),
		FullName:    fullName,
		Email:       email,
		PhoneNumber: phoneNumber,
		NationalID:  nationalID,
		Address:     address,
		DateOfBirth: dateOfBirth,
		Status:      status,
		Notes:       notes,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := b.Validate(); err != nil {
		return nil, err
	}

	return b, nil
}

// Validate checks business rules for a Beneficiary.
func (b *Beneficiary) Validate() error {
	if b.FullName == "" {
		return ErrBeneficiaryFullNameRequired
	}
	if len(b.FullName) > 255 {
		return ErrBeneficiaryFullNameTooLong
	}
	if !b.Status.IsValid() {
		return ErrBeneficiaryInvalidStatus
	}
	// Must have at least one contact method
	hasEmail := b.Email != nil && *b.Email != ""
	hasPhone := b.PhoneNumber != nil && *b.PhoneNumber != ""
	if !hasEmail && !hasPhone {
		return ErrBeneficiaryContactRequired
	}
	return nil
}

// IsValid checks if a BeneficiaryStatus value is valid.
func (s BeneficiaryStatus) IsValid() bool {
	switch s {
	case BeneficiaryStatusActive, BeneficiaryStatusInactive, BeneficiaryStatusSuspended:
		return true
	}
	return false
}

// String returns the string representation of BeneficiaryStatus.
func (s BeneficiaryStatus) String() string {
	return string(s)
}
