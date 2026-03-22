package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ProgramRepositoryMock is a test mock for repository.ProgramRepository.
type ProgramRepositoryMock struct {
	CreateFn  func(ctx context.Context, program *entity.Program) error
	GetByIDFn func(ctx context.Context, id uuid.UUID) (*entity.Program, error)
	ListFn    func(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error)
	UpdateFn  func(ctx context.Context, program *entity.Program) error
	DeleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *ProgramRepositoryMock) Create(ctx context.Context, program *entity.Program) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, program)
	}
	return nil
}

func (m *ProgramRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *ProgramRepositoryMock) List(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *ProgramRepositoryMock) Update(ctx context.Context, program *entity.Program) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, program)
	}
	return nil
}

func (m *ProgramRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
