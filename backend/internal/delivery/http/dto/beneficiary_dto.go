package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// CreateBeneficiaryRequest is the request body for creating a beneficiary.
type CreateBeneficiaryRequest struct {
	FullName    string  `json:"full_name" validate:"required,max=255"`
	Email       *string `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,max=50"`
	NationalID  *string `json:"national_id" validate:"omitempty,max=100"`
	Address     *string `json:"address" validate:"omitempty"`
	DateOfBirth *string `json:"date_of_birth" validate:"omitempty"`
	Status      string  `json:"status" validate:"required,oneof=active inactive suspended"`
	Notes       *string `json:"notes" validate:"omitempty"`
}

// UpdateBeneficiaryRequest is the request body for updating a beneficiary.
type UpdateBeneficiaryRequest struct {
	FullName    string  `json:"full_name" validate:"required,max=255"`
	Email       *string `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,max=50"`
	NationalID  *string `json:"national_id" validate:"omitempty,max=100"`
	Address     *string `json:"address" validate:"omitempty"`
	DateOfBirth *string `json:"date_of_birth" validate:"omitempty"`
	Status      string  `json:"status" validate:"required,oneof=active inactive suspended"`
	Notes       *string `json:"notes" validate:"omitempty"`
}

// EnrollProgramRequest is the request body for enrolling a beneficiary in a program.
type EnrollProgramRequest struct {
	ProgramID string `json:"program_id" validate:"required,uuid"`
}

// BeneficiaryResponse is the response body for a single beneficiary.
type BeneficiaryResponse struct {
	ID          uuid.UUID                   `json:"id"`
	FullName    string                      `json:"full_name"`
	Email       *string                     `json:"email,omitempty"`
	PhoneNumber *string                     `json:"phone_number,omitempty"`
	NationalID  *string                     `json:"national_id,omitempty"`
	Address     *string                     `json:"address,omitempty"`
	DateOfBirth *string                     `json:"date_of_birth,omitempty"`
	Status      string                      `json:"status"`
	Notes       *string                     `json:"notes,omitempty"`
	Programs    []ProgramEnrollmentResponse `json:"programs,omitempty"`
	CreatedAt   string                      `json:"created_at"`
	UpdatedAt   string                      `json:"updated_at"`
}

// ProgramEnrollmentResponse is the response for a program enrollment.
type ProgramEnrollmentResponse struct {
	ProgramID   uuid.UUID `json:"program_id"`
	ProgramName string    `json:"program_name"`
	EnrolledAt  string    `json:"enrolled_at"`
	Status      string    `json:"status"`
}

// BeneficiaryListResponse is the response body for a list of beneficiaries.
type BeneficiaryListResponse struct {
	Data   []BeneficiaryResponse `json:"data"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// ToEntity converts a CreateBeneficiaryRequest to a domain entity.
func (r *CreateBeneficiaryRequest) ToEntity() (*entity.Beneficiary, error) {
	dob, err := parseOptionalDate(r.DateOfBirth)
	if err != nil {
		return nil, err
	}

	return entity.NewBeneficiary(
		r.FullName,
		r.Email,
		r.PhoneNumber,
		r.NationalID,
		r.Address,
		dob,
		entity.BeneficiaryStatus(r.Status),
		r.Notes,
	)
}

// ToEntity converts an UpdateBeneficiaryRequest to a partial domain entity.
func (r *UpdateBeneficiaryRequest) ToEntity() (*entity.Beneficiary, error) {
	dob, err := parseOptionalDate(r.DateOfBirth)
	if err != nil {
		return nil, err
	}

	b := &entity.Beneficiary{
		FullName:    r.FullName,
		Email:       r.Email,
		PhoneNumber: r.PhoneNumber,
		NationalID:  r.NationalID,
		Address:     r.Address,
		DateOfBirth: dob,
		Status:      entity.BeneficiaryStatus(r.Status),
		Notes:       r.Notes,
	}

	return b, nil
}

// NewBeneficiaryResponse creates a BeneficiaryResponse from a domain entity.
func NewBeneficiaryResponse(b *entity.Beneficiary) BeneficiaryResponse {
	resp := BeneficiaryResponse{
		ID:          b.ID,
		FullName:    b.FullName,
		Email:       b.Email,
		PhoneNumber: b.PhoneNumber,
		NationalID:  b.NationalID,
		Address:     b.Address,
		Status:      string(b.Status),
		Notes:       b.Notes,
		CreatedAt:   b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   b.UpdatedAt.Format(time.RFC3339),
	}
	if b.DateOfBirth != nil {
		s := b.DateOfBirth.Format(dateFormat)
		resp.DateOfBirth = &s
	}
	for _, e := range b.Programs {
		resp.Programs = append(resp.Programs, ProgramEnrollmentResponse{
			ProgramID:   e.ProgramID,
			ProgramName: e.ProgramName,
			EnrolledAt:  e.EnrolledAt.Format(time.RFC3339),
			Status:      e.Status,
		})
	}
	return resp
}

// NewBeneficiaryListResponse creates a BeneficiaryListResponse from domain entities.
func NewBeneficiaryListResponse(beneficiaries []*entity.Beneficiary, total, limit, offset int) BeneficiaryListResponse {
	data := make([]BeneficiaryResponse, 0, len(beneficiaries))
	for _, b := range beneficiaries {
		data = append(data, NewBeneficiaryResponse(b))
	}
	return BeneficiaryListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
