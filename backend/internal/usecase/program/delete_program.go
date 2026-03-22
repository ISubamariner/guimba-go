package program

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeleteProgramUseCase handles soft-deleting a program.
type DeleteProgramUseCase struct {
	repo repository.ProgramRepository
}

// NewDeleteProgramUseCase creates a new DeleteProgramUseCase.
func NewDeleteProgramUseCase(repo repository.ProgramRepository) *DeleteProgramUseCase {
	return &DeleteProgramUseCase{repo: repo}
}

// Execute soft-deletes a program by ID.
func (uc *DeleteProgramUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Program", id)
	}

	return uc.repo.Delete(ctx, id)
}
