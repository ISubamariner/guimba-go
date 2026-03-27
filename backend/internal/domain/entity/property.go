package entity

import (
	"time"

	"github.com/google/uuid"
)

// Property represents a land parcel or building owned by a landlord.
type Property struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	PropertyCode       string     `json:"property_code"`
	Address            *Address   `json:"address,omitempty"`
	GeoJSONCoordinates *string    `json:"geojson_coordinates,omitempty"`
	PropertyType       string     `json:"property_type"`
	SizeInAcres        *float64   `json:"size_in_acres,omitempty"`
	SizeInSqm          float64    `json:"size_in_sqm"`
	OwnerID            uuid.UUID  `json:"owner_id"`
	IsAvailableForRent bool       `json:"is_available_for_rent"`
	IsActive           bool       `json:"is_active"`
	MonthlyRentAmount  *float64   `json:"monthly_rent_amount,omitempty"`
	Description        *string    `json:"description,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// NewProperty creates a new Property with generated ID and defaults.
func NewProperty(name, propertyCode string, address *Address, geojson *string, propertyType string, sizeInAcres *float64, sizeInSqm float64, ownerID uuid.UUID, monthlyRent *float64, description *string) (*Property, error) {
	if propertyType == "" {
		propertyType = "LAND"
	}

	p := &Property{
		ID:                 uuid.New(),
		Name:               name,
		PropertyCode:       propertyCode,
		Address:            address,
		GeoJSONCoordinates: geojson,
		PropertyType:       propertyType,
		SizeInAcres:        sizeInAcres,
		SizeInSqm:          sizeInSqm,
		OwnerID:            ownerID,
		IsAvailableForRent: true,
		IsActive:           true,
		MonthlyRentAmount:  monthlyRent,
		Description:        description,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

// Validate checks business rules for a Property.
func (p *Property) Validate() error {
	if p.Name == "" {
		return ErrPropertyNameRequired
	}
	if len(p.Name) > 255 {
		return ErrPropertyNameTooLong
	}
	if p.PropertyCode == "" {
		return ErrPropertyCodeRequired
	}
	if p.SizeInSqm <= 0 {
		return ErrPropertySizeRequired
	}
	return nil
}
