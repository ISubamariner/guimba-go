package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeleteTenantUseCase handles soft-deleting a tenant.
type DeleteTenantUseCase struct {
	repo repository.TenantRepository
}

// NewDeleteTenantUseCase creates a new DeleteTenantUseCase.
func NewDeleteTenantUseCase(repo repository.TenantRepository) *DeleteTenantUseCase {
	return &DeleteTenantUseCase{repo: repo}
}

// Execute soft-deletes a tenant by ID.
func (uc *DeleteTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	return uc.repo.Delete(ctx, id)
}
