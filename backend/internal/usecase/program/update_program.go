package program

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// UpdateProgramUseCase handles updating an existing program.
type UpdateProgramUseCase struct {
	repo repository.ProgramRepository
}

// NewUpdateProgramUseCase creates a new UpdateProgramUseCase.
func NewUpdateProgramUseCase(repo repository.ProgramRepository) *UpdateProgramUseCase {
	return &UpdateProgramUseCase{repo: repo}
}

// Execute updates a program after verifying it exists.
func (uc *UpdateProgramUseCase) Execute(ctx context.Context, id uuid.UUID, program *entity.Program) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Program", id)
	}

	program.ID = id
	program.CreatedAt = existing.CreatedAt

	if err := program.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, program)
}
