package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// DebtRepositoryMock is a test mock for repository.DebtRepository.
type DebtRepositoryMock struct {
	CreateFn                    func(ctx context.Context, debt *entity.Debt) error
	GetByIDFn                   func(ctx context.Context, id uuid.UUID) (*entity.Debt, error)
	ListFn                      func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error)
	UpdateFn                    func(ctx context.Context, debt *entity.Debt) error
	DeleteFn                    func(ctx context.Context, id uuid.UUID) error
	HasActiveDebtsForPropertyFn func(ctx context.Context, propertyID uuid.UUID) (bool, error)
}

func (m *DebtRepositoryMock) Create(ctx context.Context, debt *entity.Debt) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, debt)
	}
	return nil
}

func (m *DebtRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *DebtRepositoryMock) List(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *DebtRepositoryMock) Update(ctx context.Context, debt *entity.Debt) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, debt)
	}
	return nil
}

func (m *DebtRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *DebtRepositoryMock) HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error) {
	if m.HasActiveDebtsForPropertyFn != nil {
		return m.HasActiveDebtsForPropertyFn(ctx, propertyID)
	}
	return false, nil
}
