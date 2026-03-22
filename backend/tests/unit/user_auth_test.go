package unit

import (
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := entity.NewUser("test@example.com", "Test User", "$2a$10$hashedpassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", u.Email)
	}
	if !u.IsActive {
		t.Error("expected new user to be active")
	}
	if u.IsEmailVerified {
		t.Error("expected new user to not be email verified")
	}
}

func TestNewUser_EmptyEmail(t *testing.T) {
	_, err := entity.NewUser("", "Test", "$2a$10$hash")
	if err != entity.ErrUserEmailRequired {
		t.Errorf("expected ErrUserEmailRequired, got %v", err)
	}
}

func TestNewUser_EmptyFullName(t *testing.T) {
	_, err := entity.NewUser("a@b.com", "", "$2a$10$hash")
	if err != entity.ErrUserFullNameRequired {
		t.Errorf("expected ErrUserFullNameRequired, got %v", err)
	}
}

func TestNewUser_EmptyPassword(t *testing.T) {
	_, err := entity.NewUser("a@b.com", "Test", "")
	if err != entity.ErrUserPasswordRequired {
		t.Errorf("expected ErrUserPasswordRequired, got %v", err)
	}
}

func TestUser_HasRole(t *testing.T) {
	u := &entity.User{
		Roles: []entity.Role{
			{Name: "admin"},
			{Name: "staff"},
		},
	}

	if !u.HasRole("admin") {
		t.Error("expected HasRole('admin') to be true")
	}
	if u.HasRole("viewer") {
		t.Error("expected HasRole('viewer') to be false")
	}
}

func TestUser_HasAnyRole(t *testing.T) {
	u := &entity.User{
		Roles: []entity.Role{{Name: "staff"}},
	}

	if !u.HasAnyRole("admin", "staff") {
		t.Error("expected HasAnyRole to be true")
	}
	if u.HasAnyRole("admin", "viewer") {
		t.Error("expected HasAnyRole to be false")
	}
}

func TestUser_HasPermission(t *testing.T) {
	u := &entity.User{
		Roles: []entity.Role{
			{
				Name: "staff",
				Permissions: []entity.Permission{
					{Name: "programs.read"},
					{Name: "programs.create"},
				},
			},
		},
	}

	if !u.HasPermission("programs.read") {
		t.Error("expected HasPermission('programs.read') to be true")
	}
	if u.HasPermission("users.delete") {
		t.Error("expected HasPermission('users.delete') to be false")
	}
}

func TestRole_Validate(t *testing.T) {
	tests := []struct {
		name string
		role entity.Role
		err  error
	}{
		{"valid", entity.Role{Name: "admin", DisplayName: "Admin"}, nil},
		{"empty name", entity.Role{Name: "", DisplayName: "Admin"}, entity.ErrRoleNameRequired},
		{"empty display name", entity.Role{Name: "admin", DisplayName: ""}, entity.ErrRoleDisplayNameRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
			if err != tt.err {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
		})
	}
}

func TestHashPassword_And_Check(t *testing.T) {
	password := "securepassword123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == password {
		t.Error("hash should not equal plaintext")
	}
	if !auth.CheckPassword(password, hash) {
		t.Error("expected CheckPassword to return true for correct password")
	}
	if auth.CheckPassword("wrongpassword", hash) {
		t.Error("expected CheckPassword to return false for wrong password")
	}
}

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	manager := auth.NewJWTManager("test-secret-key-for-testing", 15*60*1e9, 7*24*60*60*1e9)

	tokens, err := manager.GenerateTokenPair(
		[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		"test@example.com",
		[]string{"admin", "staff"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	// Validate access token
	claims, err := manager.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", claims.Email)
	}
	if len(claims.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(claims.Roles))
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	manager := auth.NewJWTManager("test-secret", 15*60*1e9, 7*24*60*60*1e9)

	_, err := manager.ValidateToken("invalid.token.string")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	manager1 := auth.NewJWTManager("secret-1", 15*60*1e9, 7*24*60*60*1e9)
	manager2 := auth.NewJWTManager("secret-2", 15*60*1e9, 7*24*60*60*1e9)

	tokens, _ := manager1.GenerateTokenPair(
		[16]byte{1}, "test@example.com", []string{"admin"},
	)

	_, err := manager2.ValidateToken(tokens.AccessToken)
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}
