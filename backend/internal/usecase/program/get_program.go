package program

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// GetProgramUseCase handles retrieving a single program by ID.
type GetProgramUseCase struct {
	repo repository.ProgramRepository
}

// NewGetProgramUseCase creates a new GetProgramUseCase.
func NewGetProgramUseCase(repo repository.ProgramRepository) *GetProgramUseCase {
	return &GetProgramUseCase{repo: repo}
}

// Execute retrieves a program by ID.
func (uc *GetProgramUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
	program, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if program == nil {
		return nil, apperror.NewNotFound("Program", id)
	}
	return program, nil
}
