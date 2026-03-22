package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// RoleRepositoryMock is a test mock for repository.RoleRepository.
type RoleRepositoryMock struct {
	GetByIDFn                func(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByNameFn              func(ctx context.Context, name string) (*entity.Role, error)
	ListFn                   func(ctx context.Context) ([]*entity.Role, error)
	GetPermissionsByRoleIDFn func(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error)
}

func (m *RoleRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *RoleRepositoryMock) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	if m.GetByNameFn != nil {
		return m.GetByNameFn(ctx, name)
	}
	return nil, nil
}

func (m *RoleRepositoryMock) List(ctx context.Context) ([]*entity.Role, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *RoleRepositoryMock) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error) {
	if m.GetPermissionsByRoleIDFn != nil {
		return m.GetPermissionsByRoleIDFn(ctx, roleID)
	}
	return nil, nil
}
