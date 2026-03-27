package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// TenantRepositoryMock is a test mock for repository.TenantRepository.
type TenantRepositoryMock struct {
	CreateFn     func(ctx context.Context, tenant *entity.Tenant) error
	GetByIDFn    func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetByEmailFn func(ctx context.Context, email string) (*entity.Tenant, error)
	ListFn       func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error)
	UpdateFn     func(ctx context.Context, tenant *entity.Tenant) error
	DeleteFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *TenantRepositoryMock) Create(ctx context.Context, tenant *entity.Tenant) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tenant)
	}
	return nil
}

func (m *TenantRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *TenantRepositoryMock) GetByEmail(ctx context.Context, email string) (*entity.Tenant, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *TenantRepositoryMock) List(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *TenantRepositoryMock) Update(ctx context.Context, tenant *entity.Tenant) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, tenant)
	}
	return nil
}

func (m *TenantRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
