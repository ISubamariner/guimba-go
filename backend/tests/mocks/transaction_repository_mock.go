package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// TransactionRepositoryMock is a test mock for repository.TransactionRepository.
type TransactionRepositoryMock struct {
	CreateFn                  func(ctx context.Context, tx *entity.Transaction) error
	GetByIDFn                 func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	ListFn                    func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error)
	UpdateFn                  func(ctx context.Context, tx *entity.Transaction) error
	ExistsByReferenceNumberFn func(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error)
}

func (m *TransactionRepositoryMock) Create(ctx context.Context, tx *entity.Transaction) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tx)
	}
	return nil
}

func (m *TransactionRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *TransactionRepositoryMock) List(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *TransactionRepositoryMock) Update(ctx context.Context, tx *entity.Transaction) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, tx)
	}
	return nil
}

func (m *TransactionRepositoryMock) ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
	if m.ExistsByReferenceNumberFn != nil {
		return m.ExistsByReferenceNumberFn(ctx, debtID, refNum)
	}
	return false, nil
}
