package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// ProgramFilter holds optional filters for listing programs.
type ProgramFilter struct {
	Status *entity.ProgramStatus
	Search *string
	Limit  int
	Offset int
}

// ProgramRepository defines the interface for program persistence operations.
type ProgramRepository interface {
	Create(ctx context.Context, program *entity.Program) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Program, error)
	List(ctx context.Context, filter ProgramFilter) ([]*entity.Program, int, error)
	Update(ctx context.Context, program *entity.Program) error
	Delete(ctx context.Context, id uuid.UUID) error
}
