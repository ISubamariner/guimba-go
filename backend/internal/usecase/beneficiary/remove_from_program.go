package beneficiary

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// RemoveFromProgramUseCase handles removing a beneficiary from a program.
type RemoveFromProgramUseCase struct {
	repo repository.BeneficiaryRepository
}

// NewRemoveFromProgramUseCase creates a new RemoveFromProgramUseCase.
func NewRemoveFromProgramUseCase(repo repository.BeneficiaryRepository) *RemoveFromProgramUseCase {
	return &RemoveFromProgramUseCase{repo: repo}
}

// Execute removes a beneficiary from a program after verifying the beneficiary exists.
func (uc *RemoveFromProgramUseCase) Execute(ctx context.Context, beneficiaryID, programID uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, beneficiaryID)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Beneficiary", beneficiaryID)
	}

	return uc.repo.RemoveFromProgram(ctx, beneficiaryID, programID)
}
