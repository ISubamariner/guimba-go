package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// UpdateTenantUseCase handles updating an existing tenant.
type UpdateTenantUseCase struct {
	repo repository.TenantRepository
}

// NewUpdateTenantUseCase creates a new UpdateTenantUseCase.
func NewUpdateTenantUseCase(repo repository.TenantRepository) *UpdateTenantUseCase {
	return &UpdateTenantUseCase{repo: repo}
}

// Execute updates a tenant after verifying it exists.
func (uc *UpdateTenantUseCase) Execute(ctx context.Context, id uuid.UUID, tenant *entity.Tenant) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	tenant.ID = id
	tenant.CreatedAt = existing.CreatedAt
	tenant.LandlordID = existing.LandlordID

	if err := tenant.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, tenant)
}
