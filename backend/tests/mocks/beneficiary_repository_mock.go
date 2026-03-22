package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// BeneficiaryRepositoryMock is a test mock for repository.BeneficiaryRepository.
type BeneficiaryRepositoryMock struct {
	CreateFn            func(ctx context.Context, beneficiary *entity.Beneficiary) error
	GetByIDFn           func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error)
	ListFn              func(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error)
	UpdateFn            func(ctx context.Context, beneficiary *entity.Beneficiary) error
	DeleteFn            func(ctx context.Context, id uuid.UUID) error
	EnrollInProgramFn   func(ctx context.Context, beneficiaryID, programID uuid.UUID) error
	RemoveFromProgramFn func(ctx context.Context, beneficiaryID, programID uuid.UUID) error
}

func (m *BeneficiaryRepositoryMock) Create(ctx context.Context, beneficiary *entity.Beneficiary) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, beneficiary)
	}
	return nil
}

func (m *BeneficiaryRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *BeneficiaryRepositoryMock) List(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *BeneficiaryRepositoryMock) Update(ctx context.Context, beneficiary *entity.Beneficiary) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, beneficiary)
	}
	return nil
}

func (m *BeneficiaryRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *BeneficiaryRepositoryMock) EnrollInProgram(ctx context.Context, beneficiaryID, programID uuid.UUID) error {
	if m.EnrollInProgramFn != nil {
		return m.EnrollInProgramFn(ctx, beneficiaryID, programID)
	}
	return nil
}

func (m *BeneficiaryRepositoryMock) RemoveFromProgram(ctx context.Context, beneficiaryID, programID uuid.UUID) error {
	if m.RemoveFromProgramFn != nil {
		return m.RemoveFromProgramFn(ctx, beneficiaryID, programID)
	}
	return nil
}
