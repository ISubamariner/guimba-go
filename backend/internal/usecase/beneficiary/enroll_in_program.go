package beneficiary

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// EnrollInProgramUseCase handles enrolling a beneficiary in a program.
type EnrollInProgramUseCase struct {
	beneficiaryRepo repository.BeneficiaryRepository
	programRepo     repository.ProgramRepository
}

// NewEnrollInProgramUseCase creates a new EnrollInProgramUseCase.
func NewEnrollInProgramUseCase(beneficiaryRepo repository.BeneficiaryRepository, programRepo repository.ProgramRepository) *EnrollInProgramUseCase {
	return &EnrollInProgramUseCase{
		beneficiaryRepo: beneficiaryRepo,
		programRepo:     programRepo,
	}
}

// Execute enrolls a beneficiary in a program after verifying both exist.
func (uc *EnrollInProgramUseCase) Execute(ctx context.Context, beneficiaryID, programID uuid.UUID) error {
	beneficiary, err := uc.beneficiaryRepo.GetByID(ctx, beneficiaryID)
	if err != nil {
		return err
	}
	if beneficiary == nil {
		return apperror.NewNotFound("Beneficiary", beneficiaryID)
	}

	program, err := uc.programRepo.GetByID(ctx, programID)
	if err != nil {
		return err
	}
	if program == nil {
		return apperror.NewNotFound("Program", programID)
	}

	return uc.beneficiaryRepo.EnrollInProgram(ctx, beneficiaryID, programID)
}
