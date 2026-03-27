package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeactivateTenantUseCase handles deactivating a tenant.
type DeactivateTenantUseCase struct {
	repo repository.TenantRepository
}

// NewDeactivateTenantUseCase creates a new DeactivateTenantUseCase.
func NewDeactivateTenantUseCase(repo repository.TenantRepository) *DeactivateTenantUseCase {
	return &DeactivateTenantUseCase{repo: repo}
}

// Execute deactivates a tenant by setting IsActive to false.
func (uc *DeactivateTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	existing.IsActive = false
	return uc.repo.Update(ctx, existing)
}
