package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// PropertyFilter holds optional filters for listing properties.
type PropertyFilter struct {
	OwnerID            *uuid.UUID
	IsActive           *bool
	IsAvailableForRent *bool
	PropertyType       *string
	Search             *string
	Limit              int
	Offset             int
}

// PropertyRepository defines the interface for property persistence operations.
type PropertyRepository interface {
	Create(ctx context.Context, property *entity.Property) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Property, error)
	GetByPropertyCode(ctx context.Context, code string) (*entity.Property, error)
	List(ctx context.Context, filter PropertyFilter) ([]*entity.Property, int, error)
	Update(ctx context.Context, property *entity.Property) error
	Delete(ctx context.Context, id uuid.UUID) error
}
