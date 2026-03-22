package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	authuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/auth"
	useruc "github.com/ISubamariner/guimba-go/backend/internal/usecase/user"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTestJWT() *auth.JWTManager {
	return auth.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
}

func TestRegisterUseCase_Success(t *testing.T) {
	viewerRole := &entity.Role{ID: uuid.New(), Name: "viewer", DisplayName: "Viewer"}

	userMock := &mocks.UserRepositoryMock{
		GetByEmailFn: func(ctx context.Context, email string) (*entity.User, error) { return nil, nil },
		CreateFn:     func(ctx context.Context, u *entity.User) error { return nil },
		AssignRoleFn: func(ctx context.Context, userID, roleID uuid.UUID) error { return nil },
	}
	roleMock := &mocks.RoleRepositoryMock{
		GetByNameFn: func(ctx context.Context, name string) (*entity.Role, error) { return viewerRole, nil },
	}

	uc := authuc.NewRegisterUseCase(userMock, roleMock, newTestJWT())
	user, tokens, err := uc.Execute(context.Background(), "test@example.com", "Test User", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", user.Email)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestRegisterUseCase_DuplicateEmail(t *testing.T) {
	existing := &entity.User{Email: "taken@example.com"}
	userMock := &mocks.UserRepositoryMock{
		GetByEmailFn: func(ctx context.Context, email string) (*entity.User, error) { return existing, nil },
	}
	roleMock := &mocks.RoleRepositoryMock{}

	uc := authuc.NewRegisterUseCase(userMock, roleMock, newTestJWT())
	_, _, err := uc.Execute(context.Background(), "taken@example.com", "Test", "password123")

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeConflict {
		t.Errorf("expected CONFLICT error, got %v", err)
	}
}

func TestLoginUseCase_Success(t *testing.T) {
	hashed, _ := auth.HashPassword("correctpassword")
	existingUser := &entity.User{
		ID: uuid.New(), Email: "user@example.com", FullName: "User",
		HashedPassword: hashed, IsActive: true,
		Roles:     []entity.Role{{Name: "viewer"}},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	userMock := &mocks.UserRepositoryMock{
		GetByEmailFn:      func(ctx context.Context, email string) (*entity.User, error) { return existingUser, nil },
		UpdateLastLoginFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}

	uc := authuc.NewLoginUseCase(userMock, newTestJWT())
	user, tokens, err := uc.Execute(context.Background(), "user@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email, got %q", user.Email)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	hashed, _ := auth.HashPassword("correctpassword")
	existingUser := &entity.User{
		ID: uuid.New(), Email: "user@example.com", HashedPassword: hashed, IsActive: true,
	}

	userMock := &mocks.UserRepositoryMock{
		GetByEmailFn: func(ctx context.Context, email string) (*entity.User, error) { return existingUser, nil },
	}

	uc := authuc.NewLoginUseCase(userMock, newTestJWT())
	_, _, err := uc.Execute(context.Background(), "user@example.com", "wrongpassword")

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

func TestLoginUseCase_UserNotFound(t *testing.T) {
	userMock := &mocks.UserRepositoryMock{
		GetByEmailFn: func(ctx context.Context, email string) (*entity.User, error) { return nil, nil },
	}

	uc := authuc.NewLoginUseCase(userMock, newTestJWT())
	_, _, err := uc.Execute(context.Background(), "nobody@example.com", "password")

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

func TestLoginUseCase_InactiveUser(t *testing.T) {
	hashed, _ := auth.HashPassword("password")
	inactiveUser := &entity.User{
		ID: uuid.New(), Email: "inactive@example.com", HashedPassword: hashed, IsActive: false,
	}

	userMock := &mocks.UserRepositoryMock{
		GetByEmailFn: func(ctx context.Context, email string) (*entity.User, error) { return inactiveUser, nil },
	}

	uc := authuc.NewLoginUseCase(userMock, newTestJWT())
	_, _, err := uc.Execute(context.Background(), "inactive@example.com", "password")

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeForbidden {
		t.Errorf("expected FORBIDDEN error, got %v", err)
	}
}

func TestGetProfileUseCase_Success(t *testing.T) {
	id := uuid.New()
	existingUser := &entity.User{ID: id, Email: "user@example.com", FullName: "User", HashedPassword: "hash", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	userMock := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, gotID uuid.UUID) (*entity.User, error) { return existingUser, nil },
	}

	uc := authuc.NewGetProfileUseCase(userMock)
	user, err := uc.Execute(context.Background(), id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != id {
		t.Errorf("expected ID %v, got %v", id, user.ID)
	}
}

func TestGetProfileUseCase_NotFound(t *testing.T) {
	userMock := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) { return nil, nil },
	}

	uc := authuc.NewGetProfileUseCase(userMock)
	_, err := uc.Execute(context.Background(), uuid.New())

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NOT_FOUND error, got %v", err)
	}
}

func TestListUsersUseCase_DefaultPagination(t *testing.T) {
	userMock := &mocks.UserRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int, error) {
			if filter.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", filter.Limit)
			}
			return nil, 0, nil
		},
	}

	uc := useruc.NewListUsersUseCase(userMock)
	_, _, err := uc.Execute(context.Background(), repository.UserFilter{Limit: 0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteUserUseCase_NotFound(t *testing.T) {
	userMock := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) { return nil, nil },
	}

	uc := useruc.NewDeleteUserUseCase(userMock)
	err := uc.Execute(context.Background(), uuid.New())

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NOT_FOUND error, got %v", err)
	}
}

func TestAssignRoleUseCase_Success(t *testing.T) {
	userID, roleID := uuid.New(), uuid.New()
	existingUser := &entity.User{ID: userID, Email: "u@b.com", FullName: "U", HashedPassword: "h", IsActive: true}
	existingRole := &entity.Role{ID: roleID, Name: "staff", DisplayName: "Staff"}

	userMock := &mocks.UserRepositoryMock{
		GetByIDFn:    func(ctx context.Context, id uuid.UUID) (*entity.User, error) { return existingUser, nil },
		AssignRoleFn: func(ctx context.Context, uid, rid uuid.UUID) error { return nil },
	}
	roleMock := &mocks.RoleRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Role, error) { return existingRole, nil },
	}

	uc := useruc.NewAssignRoleUseCase(userMock, roleMock)
	err := uc.Execute(context.Background(), userID, roleID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
