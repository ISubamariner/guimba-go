package beneficiary

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ListBeneficiariesUseCase handles listing beneficiaries with filtering and pagination.
type ListBeneficiariesUseCase struct {
	repo repository.BeneficiaryRepository
}

// NewListBeneficiariesUseCase creates a new ListBeneficiariesUseCase.
func NewListBeneficiariesUseCase(repo repository.BeneficiaryRepository) *ListBeneficiariesUseCase {
	return &ListBeneficiariesUseCase{repo: repo}
}

// Execute returns a filtered, paginated list of beneficiaries and the total count.
func (uc *ListBeneficiariesUseCase) Execute(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
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
