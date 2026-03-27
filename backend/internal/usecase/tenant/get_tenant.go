package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// GetTenantUseCase handles retrieving a single tenant by ID.
type GetTenantUseCase struct {
	repo repository.TenantRepository
}

// NewGetTenantUseCase creates a new GetTenantUseCase.
func NewGetTenantUseCase(repo repository.TenantRepository) *GetTenantUseCase {
	return &GetTenantUseCase{repo: repo}
}

// Execute retrieves a tenant by ID.
func (uc *GetTenantUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, apperror.NewNotFound("Tenant", id)
	}
	return tenant, nil
}
