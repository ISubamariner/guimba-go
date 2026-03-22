package program

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ListProgramsUseCase handles listing programs with filtering and pagination.
type ListProgramsUseCase struct {
	repo repository.ProgramRepository
}

// NewListProgramsUseCase creates a new ListProgramsUseCase.
func NewListProgramsUseCase(repo repository.ProgramRepository) *ListProgramsUseCase {
	return &ListProgramsUseCase{repo: repo}
}

// Execute returns a filtered, paginated list of programs and the total count.
func (uc *ListProgramsUseCase) Execute(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return uc.repo.List(ctx, filter)
}
