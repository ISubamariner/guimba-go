package program

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// CreateProgramUseCase handles creating a new program.
type CreateProgramUseCase struct {
	repo repository.ProgramRepository
}

// NewCreateProgramUseCase creates a new CreateProgramUseCase.
func NewCreateProgramUseCase(repo repository.ProgramRepository) *CreateProgramUseCase {
	return &CreateProgramUseCase{repo: repo}
}

// Execute creates a new program and persists it.
func (uc *CreateProgramUseCase) Execute(ctx context.Context, program *entity.Program) error {
	if err := program.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, program)
}
