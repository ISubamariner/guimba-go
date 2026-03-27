package tenant

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ListTenantsUseCase handles listing tenants with filtering and pagination.
type ListTenantsUseCase struct {
	repo repository.TenantRepository
}

// NewListTenantsUseCase creates a new ListTenantsUseCase.
func NewListTenantsUseCase(repo repository.TenantRepository) *ListTenantsUseCase {
	return &ListTenantsUseCase{repo: repo}
}

// Execute returns a filtered, paginated list of tenants and the total count.
func (uc *ListTenantsUseCase) Execute(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return uc.repo.List(ctx, filter)
}
