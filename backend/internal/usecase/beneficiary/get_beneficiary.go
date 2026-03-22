package beneficiary

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// GetBeneficiaryUseCase handles retrieving a single beneficiary by ID.
type GetBeneficiaryUseCase struct {
	repo repository.BeneficiaryRepository
}

// NewGetBeneficiaryUseCase creates a new GetBeneficiaryUseCase.
func NewGetBeneficiaryUseCase(repo repository.BeneficiaryRepository) *GetBeneficiaryUseCase {
	return &GetBeneficiaryUseCase{repo: repo}
}

// Execute retrieves a beneficiary by ID.
func (uc *GetBeneficiaryUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) {
	beneficiary, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if beneficiary == nil {
		return nil, apperror.NewNotFound("Beneficiary", id)
	}
	return beneficiary, nil
}
