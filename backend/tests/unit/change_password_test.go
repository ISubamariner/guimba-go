package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	authuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/auth"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestChangePasswordUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	currentPass := "OldPassword123"
	newPass := "NewPassword456"

	// Hash the current password for the test user
	currentHashed, _ := auth.HashPassword(currentPass)

	tests := []struct {
		name           string
		userID         uuid.UUID
		currentPass    string
		newPass        string
		mockGetByID    func(ctx context.Context, id uuid.UUID) (*entity.User, error)
		mockUpdatePass func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
		expectError    bool
		errorCode      apperror.Code
	}{
		{
			name:        "success - password changed",
			userID:      userID,
			currentPass: currentPass,
			newPass:     newPass,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
				return &entity.User{
					ID:             userID,
					Email:          "test@example.com",
					HashedPassword: currentHashed,
				}, nil
			},
			mockUpdatePass: func(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
				return nil
			},
			expectError: false,
		},
		{
			name:        "failure - wrong current password",
			userID:      userID,
			currentPass: "WrongPassword",
			newPass:     newPass,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
				return &entity.User{
					ID:             userID,
					Email:          "test@example.com",
					HashedPassword: currentHashed,
				}, nil
			},
			mockUpdatePass: nil,
			expectError:    true,
			errorCode:      apperror.CodeValidation,
		},
		{
			name:        "failure - new password same as current",
			userID:      userID,
			currentPass: currentPass,
			newPass:     currentPass,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
				return &entity.User{
					ID:             userID,
					Email:          "test@example.com",
					HashedPassword: currentHashed,
				}, nil
			},
			mockUpdatePass: nil,
			expectError:    true,
			errorCode:      apperror.CodeValidation,
		},
		{
			name:        "failure - user not found",
			userID:      userID,
			currentPass: currentPass,
			newPass:     newPass,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
				return nil, nil
			},
			mockUpdatePass: nil,
			expectError:    true,
			errorCode:      apperror.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.UserRepositoryMock{
				GetByIDFn:        tt.mockGetByID,
				UpdatePasswordFn: tt.mockUpdatePass,
			}

			uc := authuc.NewChangePasswordUseCase(mockRepo)
			err := uc.Execute(ctx, tt.userID, tt.currentPass, tt.newPass)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if appErr, ok := err.(*apperror.AppError); ok {
					if appErr.Code != tt.errorCode {
						t.Errorf("expected error code %s, got %s", tt.errorCode, appErr.Code)
					}
				} else {
					t.Errorf("expected AppError, got %T", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}
