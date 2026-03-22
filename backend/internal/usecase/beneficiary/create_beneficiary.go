package beneficiary

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// CreateBeneficiaryUseCase handles creating a new beneficiary.
type CreateBeneficiaryUseCase struct {
	repo repository.BeneficiaryRepository
}

// NewCreateBeneficiaryUseCase creates a new CreateBeneficiaryUseCase.
func NewCreateBeneficiaryUseCase(repo repository.BeneficiaryRepository) *CreateBeneficiaryUseCase {
	return &CreateBeneficiaryUseCase{repo: repo}
}

// Execute creates a new beneficiary and persists it.
func (uc *CreateBeneficiaryUseCase) Execute(ctx context.Context, beneficiary *entity.Beneficiary) error {
	if err := beneficiary.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, beneficiary)
}
