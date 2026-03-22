package beneficiary

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// UpdateBeneficiaryUseCase handles updating an existing beneficiary.
type UpdateBeneficiaryUseCase struct {
	repo repository.BeneficiaryRepository
}

// NewUpdateBeneficiaryUseCase creates a new UpdateBeneficiaryUseCase.
func NewUpdateBeneficiaryUseCase(repo repository.BeneficiaryRepository) *UpdateBeneficiaryUseCase {
	return &UpdateBeneficiaryUseCase{repo: repo}
}

// Execute updates a beneficiary after verifying it exists.
func (uc *UpdateBeneficiaryUseCase) Execute(ctx context.Context, id uuid.UUID, beneficiary *entity.Beneficiary) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Beneficiary", id)
	}

	beneficiary.ID = id
	beneficiary.CreatedAt = existing.CreatedAt

	if err := beneficiary.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, beneficiary)
}
