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

// createTestLandlord creates a user and returns its ID for use as a landlord.
func createTestLandlord(t *testing.T) uuid.UUID {
	t.Helper()
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	email := "landlord-" + uuid.New().String() + "@example.com"
	user := newTestUser(email)
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create landlord user: %v", err)
	}
	return user.ID
}

func newTestTenant(name string, landlordID uuid.UUID) *entity.Tenant {
	now := time.Now().UTC()
	email := name + "@example.com"
	phone := "+1234567890"
	return &entity.Tenant{
		ID:          uuid.New(),
		FullName:    name,
		Email:       &email,
		PhoneNumber: &phone,
		LandlordID:  landlordID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestTenantRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlordID := createTestLandlord(t)
	repo := pg.NewTenantRepoPG(testPgPool)

	tenant := newTestTenant("Test Tenant", landlordID)
	err := repo.Create(ctx, tenant)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByID
	found, err := repo.GetByID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected tenant, got nil")
	}
	if found.FullName != tenant.FullName {
		t.Errorf("expected full_name %q, got %q", tenant.FullName, found.FullName)
	}
	if found.LandlordID != landlordID {
		t.Errorf("expected landlord_id %v, got %v", landlordID, found.LandlordID)
	}
}

func TestTenantRepo_List_ScopedByLandlord(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlord1 := createTestLandlord(t)
	landlord2 := createTestLandlord(t)

	repo := pg.NewTenantRepoPG(testPgPool)

	// Create 2 tenants for landlord1
	for i := 0; i < 2; i++ {
		tenant := newTestTenant("Landlord1-Tenant-"+string(rune('A'+i)), landlord1)
		if err := repo.Create(ctx, tenant); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Create 1 tenant for landlord2
	tenant := newTestTenant("Landlord2-Tenant", landlord2)
	if err := repo.Create(ctx, tenant); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List tenants for landlord1
	tenants, total, err := repo.List(ctx, repository.TenantFilter{
		LandlordID: &landlord1,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 tenants for landlord1, got %d", total)
	}
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(tenants))
	}

	// List tenants for landlord2
	tenants, total, err = repo.List(ctx, repository.TenantFilter{
		LandlordID: &landlord2,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 tenant for landlord2, got %d", total)
	}
	if len(tenants) != 1 {
		t.Errorf("expected 1 tenant, got %d", len(tenants))
	}
}
