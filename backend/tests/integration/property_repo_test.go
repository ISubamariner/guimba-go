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

func newTestProperty(name string, landlordID uuid.UUID) *entity.Property {
	now := time.Now().UTC()
	propertyCode := "PROP-" + uuid.New().String()[:8]
	return &entity.Property{
		ID:                 uuid.New(),
		Name:               name,
		PropertyCode:       propertyCode,
		PropertyType:       "LAND",
		SizeInSqm:          100.0,
		OwnerID:            landlordID,
		IsAvailableForRent: true,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func TestPropertyRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlordID := createTestLandlord(t)
	repo := pg.NewPropertyRepoPG(testPgPool)

	property := newTestProperty("Test Property", landlordID)
	err := repo.Create(ctx, property)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByID
	found, err := repo.GetByID(ctx, property.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected property, got nil")
	}
	if found.Name != property.Name {
		t.Errorf("expected name %q, got %q", property.Name, found.Name)
	}
	if found.OwnerID != landlordID {
		t.Errorf("expected owner_id %v, got %v", landlordID, found.OwnerID)
	}
	if found.PropertyCode != property.PropertyCode {
		t.Errorf("expected property_code %q, got %q", property.PropertyCode, found.PropertyCode)
	}
}

func TestPropertyRepo_List_ScopedByLandlord(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlord1 := createTestLandlord(t)
	landlord2 := createTestLandlord(t)

	repo := pg.NewPropertyRepoPG(testPgPool)

	// Create 2 properties for landlord1
	for i := 0; i < 2; i++ {
		property := newTestProperty("Landlord1-Property-"+string(rune('A'+i)), landlord1)
		if err := repo.Create(ctx, property); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Create 1 property for landlord2
	property := newTestProperty("Landlord2-Property", landlord2)
	if err := repo.Create(ctx, property); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List properties for landlord1
	properties, total, err := repo.List(ctx, repository.PropertyFilter{
		OwnerID: &landlord1,
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 properties for landlord1, got %d", total)
	}
	if len(properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(properties))
	}

	// List properties for landlord2
	properties, total, err = repo.List(ctx, repository.PropertyFilter{
		OwnerID: &landlord2,
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 property for landlord2, got %d", total)
	}
	if len(properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(properties))
	}
}
