package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// TenantFilter holds optional filters for listing tenants.
type TenantFilter struct {
	LandlordID *uuid.UUID
	IsActive   *bool
	Search     *string
	Limit      int
	Offset     int
}

// TenantRepository defines the interface for tenant persistence operations.
type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetByEmail(ctx context.Context, email string) (*entity.Tenant, error)
	List(ctx context.Context, filter TenantFilter) ([]*entity.Tenant, int, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}
