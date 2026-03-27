package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// AddressDTO is the request/response shape for an address.
type AddressDTO struct {
	Street        string `json:"street" validate:"required"`
	City          string `json:"city" validate:"required"`
	StateOrRegion string `json:"state_or_region" validate:"required"`
	PostalCode    string `json:"postal_code,omitempty" validate:"omitempty"`
	Country       string `json:"country,omitempty" validate:"omitempty"`
}

// CreateTenantRequest is the request body for creating a tenant.
type CreateTenantRequest struct {
	FullName    string      `json:"full_name" validate:"required,max=255"`
	Email       *string     `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber *string     `json:"phone_number" validate:"omitempty,max=50"`
	NationalID  *string     `json:"national_id" validate:"omitempty,max=100"`
	Address     *AddressDTO `json:"address" validate:"omitempty"`
	Notes       *string     `json:"notes" validate:"omitempty"`
}

// UpdateTenantRequest is the request body for updating a tenant.
type UpdateTenantRequest struct {
	FullName    string      `json:"full_name" validate:"required,max=255"`
	Email       *string     `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber *string     `json:"phone_number" validate:"omitempty,max=50"`
	NationalID  *string     `json:"national_id" validate:"omitempty,max=100"`
	Address     *AddressDTO `json:"address" validate:"omitempty"`
	Notes       *string     `json:"notes" validate:"omitempty"`
}

// TenantResponse is the response body for a single tenant.
type TenantResponse struct {
	ID          uuid.UUID   `json:"id"`
	FullName    string      `json:"full_name"`
	Email       *string     `json:"email,omitempty"`
	PhoneNumber *string     `json:"phone_number,omitempty"`
	NationalID  *string     `json:"national_id,omitempty"`
	Address     *AddressDTO `json:"address,omitempty"`
	LandlordID  uuid.UUID   `json:"landlord_id"`
	IsActive    bool        `json:"is_active"`
	Notes       *string     `json:"notes,omitempty"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// TenantListResponse is the response body for a list of tenants.
type TenantListResponse struct {
	Data   []TenantResponse `json:"data"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// ToEntity converts a CreateTenantRequest to a domain entity.
func (r *CreateTenantRequest) ToEntity(landlordID uuid.UUID) (*entity.Tenant, error) {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	return entity.NewTenant(r.FullName, r.Email, r.PhoneNumber, r.NationalID, addr, landlordID, r.Notes)
}

// ToEntity converts an UpdateTenantRequest to a partial domain entity.
func (r *UpdateTenantRequest) ToEntity() *entity.Tenant {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	return &entity.Tenant{
		FullName:    r.FullName,
		Email:       r.Email,
		PhoneNumber: r.PhoneNumber,
		NationalID:  r.NationalID,
		Address:     addr,
		Notes:       r.Notes,
	}
}

// NewTenantResponse creates a TenantResponse from a domain entity.
func NewTenantResponse(t *entity.Tenant) TenantResponse {
	resp := TenantResponse{
		ID:          t.ID,
		FullName:    t.FullName,
		Email:       t.Email,
		PhoneNumber: t.PhoneNumber,
		NationalID:  t.NationalID,
		LandlordID:  t.LandlordID,
		IsActive:    t.IsActive,
		Notes:       t.Notes,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
	if t.Address != nil {
		resp.Address = &AddressDTO{
			Street:        t.Address.Street,
			City:          t.Address.City,
			StateOrRegion: t.Address.StateOrRegion,
			PostalCode:    t.Address.PostalCode,
			Country:       t.Address.Country,
		}
	}
	return resp
}

// NewTenantListResponse creates a TenantListResponse from domain entities.
func NewTenantListResponse(tenants []*entity.Tenant, total, limit, offset int) TenantListResponse {
	data := make([]TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		data = append(data, NewTenantResponse(t))
	}
	return TenantListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
