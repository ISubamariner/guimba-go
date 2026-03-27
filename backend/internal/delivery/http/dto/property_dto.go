package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// CreatePropertyRequest is the request body for creating a property.
type CreatePropertyRequest struct {
	Name               string      `json:"name" validate:"required,max=255"`
	PropertyCode       string      `json:"property_code" validate:"required,max=100"`
	Address            *AddressDTO `json:"address" validate:"omitempty"`
	GeoJSONCoordinates *string     `json:"geojson_coordinates" validate:"omitempty"`
	PropertyType       string      `json:"property_type" validate:"omitempty,max=50"`
	SizeInAcres        *float64    `json:"size_in_acres" validate:"omitempty,gt=0"`
	SizeInSqm          float64     `json:"size_in_sqm" validate:"required,gt=0"`
	MonthlyRentAmount  *float64    `json:"monthly_rent_amount" validate:"omitempty,gte=0"`
	Description        *string     `json:"description" validate:"omitempty"`
}

// UpdatePropertyRequest is the request body for updating a property.
type UpdatePropertyRequest struct {
	Name               string      `json:"name" validate:"required,max=255"`
	PropertyCode       string      `json:"property_code" validate:"required,max=100"`
	Address            *AddressDTO `json:"address" validate:"omitempty"`
	GeoJSONCoordinates *string     `json:"geojson_coordinates" validate:"omitempty"`
	PropertyType       string      `json:"property_type" validate:"omitempty,max=50"`
	SizeInAcres        *float64    `json:"size_in_acres" validate:"omitempty,gt=0"`
	SizeInSqm          float64     `json:"size_in_sqm" validate:"required,gt=0"`
	IsAvailableForRent *bool       `json:"is_available_for_rent" validate:"omitempty"`
	MonthlyRentAmount  *float64    `json:"monthly_rent_amount" validate:"omitempty,gte=0"`
	Description        *string     `json:"description" validate:"omitempty"`
}

// PropertyResponse is the response body for a single property.
type PropertyResponse struct {
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	PropertyCode       string      `json:"property_code"`
	Address            *AddressDTO `json:"address,omitempty"`
	GeoJSONCoordinates *string     `json:"geojson_coordinates,omitempty"`
	PropertyType       string      `json:"property_type"`
	SizeInAcres        *float64    `json:"size_in_acres,omitempty"`
	SizeInSqm          float64     `json:"size_in_sqm"`
	OwnerID            uuid.UUID   `json:"owner_id"`
	IsAvailableForRent bool        `json:"is_available_for_rent"`
	IsActive           bool        `json:"is_active"`
	MonthlyRentAmount  *float64    `json:"monthly_rent_amount,omitempty"`
	Description        *string     `json:"description,omitempty"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
}

// PropertyListResponse is the response body for a list of properties.
type PropertyListResponse struct {
	Data   []PropertyResponse `json:"data"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// ToEntity converts a CreatePropertyRequest to a domain entity.
func (r *CreatePropertyRequest) ToEntity(ownerID uuid.UUID) (*entity.Property, error) {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	return entity.NewProperty(r.Name, r.PropertyCode, addr, r.GeoJSONCoordinates, r.PropertyType, r.SizeInAcres, r.SizeInSqm, ownerID, r.MonthlyRentAmount, r.Description)
}

// ToEntity converts an UpdatePropertyRequest to a partial domain entity.
func (r *UpdatePropertyRequest) ToEntity() *entity.Property {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	p := &entity.Property{
		Name:               r.Name,
		PropertyCode:       r.PropertyCode,
		Address:            addr,
		GeoJSONCoordinates: r.GeoJSONCoordinates,
		PropertyType:       r.PropertyType,
		SizeInAcres:        r.SizeInAcres,
		SizeInSqm:          r.SizeInSqm,
		IsAvailableForRent: true,
		MonthlyRentAmount:  r.MonthlyRentAmount,
		Description:        r.Description,
	}
	if r.IsAvailableForRent != nil {
		p.IsAvailableForRent = *r.IsAvailableForRent
	}
	return p
}

// NewPropertyResponse creates a PropertyResponse from a domain entity.
func NewPropertyResponse(p *entity.Property) PropertyResponse {
	resp := PropertyResponse{
		ID:                 p.ID,
		Name:               p.Name,
		PropertyCode:       p.PropertyCode,
		GeoJSONCoordinates: p.GeoJSONCoordinates,
		PropertyType:       p.PropertyType,
		SizeInAcres:        p.SizeInAcres,
		SizeInSqm:          p.SizeInSqm,
		OwnerID:            p.OwnerID,
		IsAvailableForRent: p.IsAvailableForRent,
		IsActive:           p.IsActive,
		MonthlyRentAmount:  p.MonthlyRentAmount,
		Description:        p.Description,
		CreatedAt:          p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          p.UpdatedAt.Format(time.RFC3339),
	}
	if p.Address != nil {
		resp.Address = &AddressDTO{
			Street:        p.Address.Street,
			City:          p.Address.City,
			StateOrRegion: p.Address.StateOrRegion,
			PostalCode:    p.Address.PostalCode,
			Country:       p.Address.Country,
		}
	}
	return resp
}

// NewPropertyListResponse creates a PropertyListResponse from domain entities.
func NewPropertyListResponse(properties []*entity.Property, total, limit, offset int) PropertyListResponse {
	data := make([]PropertyResponse, 0, len(properties))
	for _, p := range properties {
		data = append(data, NewPropertyResponse(p))
	}
	return PropertyListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
