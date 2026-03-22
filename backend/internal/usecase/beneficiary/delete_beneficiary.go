package beneficiary

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeleteBeneficiaryUseCase handles soft-deleting a beneficiary.
type DeleteBeneficiaryUseCase struct {
	repo repository.BeneficiaryRepository
}

// NewDeleteBeneficiaryUseCase creates a new DeleteBeneficiaryUseCase.
func NewDeleteBeneficiaryUseCase(repo repository.BeneficiaryRepository) *DeleteBeneficiaryUseCase {
	return &DeleteBeneficiaryUseCase{repo: repo}
}

// Execute soft-deletes a beneficiary by ID.
func (uc *DeleteBeneficiaryUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Beneficiary", id)
	}

	return uc.repo.Delete(ctx, id)
}
