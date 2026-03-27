package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// PropertyRepositoryMock is a test mock for repository.PropertyRepository.
type PropertyRepositoryMock struct {
	CreateFn            func(ctx context.Context, property *entity.Property) error
	GetByIDFn           func(ctx context.Context, id uuid.UUID) (*entity.Property, error)
	GetByPropertyCodeFn func(ctx context.Context, code string) (*entity.Property, error)
	ListFn              func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error)
	UpdateFn            func(ctx context.Context, property *entity.Property) error
	DeleteFn            func(ctx context.Context, id uuid.UUID) error
}

func (m *PropertyRepositoryMock) Create(ctx context.Context, property *entity.Property) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, property)
	}
	return nil
}

func (m *PropertyRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *PropertyRepositoryMock) GetByPropertyCode(ctx context.Context, code string) (*entity.Property, error) {
	if m.GetByPropertyCodeFn != nil {
		return m.GetByPropertyCodeFn(ctx, code)
	}
	return nil, nil
}

func (m *PropertyRepositoryMock) List(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *PropertyRepositoryMock) Update(ctx context.Context, property *entity.Property) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, property)
	}
	return nil
}

func (m *PropertyRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
