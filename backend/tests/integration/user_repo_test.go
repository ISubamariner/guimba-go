//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
)

func newTestUser(email string) *entity.User {
	now := time.Now().UTC()
	return &entity.User{
		ID:              uuid.New(),
		Email:           email,
		FullName:        "Test User",
		HashedPassword:  "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGH", // bcrypt hash
		IsActive:        true,
		IsEmailVerified: false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func TestUserRepo_Create_And_GetByEmail(t *testing.T) {
	truncateAll(t)
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	user := newTestUser("test@example.com")
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByEmail
	found, err := repo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, found.Email)
	}
	if found.FullName != user.FullName {
		t.Errorf("expected full_name %q, got %q", user.FullName, found.FullName)
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	truncateAll(t)
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	user := newTestUser("getbyid@example.com")
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByID
	found, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.ID != user.ID {
		t.Errorf("expected ID %v, got %v", user.ID, found.ID)
	}
	if found.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, found.Email)
	}
}

func TestUserRepo_List(t *testing.T) {
	truncateAll(t)
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	// Create 3 users
	for i := 0; i < 3; i++ {
		email := string(rune('a'+i)) + "@example.com"
		if err := repo.Create(ctx, newTestUser(email)); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	users, total, err := repo.List(ctx, repository.UserFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}
