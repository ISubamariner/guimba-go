package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// UserRepositoryMock is a test mock for repository.UserRepository.
type UserRepositoryMock struct {
	CreateFn          func(ctx context.Context, user *entity.User) error
	GetByIDFn         func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmailFn      func(ctx context.Context, email string) (*entity.User, error)
	ListFn            func(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int, error)
	UpdateFn          func(ctx context.Context, user *entity.User) error
	DeleteFn          func(ctx context.Context, id uuid.UUID) error
	AssignRoleFn      func(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFn      func(ctx context.Context, userID, roleID uuid.UUID) error
	UpdateLastLoginFn func(ctx context.Context, id uuid.UUID) error
	UpdatePasswordFn  func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}

func (m *UserRepositoryMock) Create(ctx context.Context, user *entity.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *UserRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *UserRepositoryMock) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *UserRepositoryMock) List(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *UserRepositoryMock) Update(ctx context.Context, user *entity.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

func (m *UserRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *UserRepositoryMock) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.AssignRoleFn != nil {
		return m.AssignRoleFn(ctx, userID, roleID)
	}
	return nil
}

func (m *UserRepositoryMock) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.RemoveRoleFn != nil {
		return m.RemoveRoleFn(ctx, userID, roleID)
	}
	return nil
}

func (m *UserRepositoryMock) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	if m.UpdateLastLoginFn != nil {
		return m.UpdateLastLoginFn(ctx, id)
	}
	return nil
}

func (m *UserRepositoryMock) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(ctx, userID, hashedPassword)
	}
	return nil
}
