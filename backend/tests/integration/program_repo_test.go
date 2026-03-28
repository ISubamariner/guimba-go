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

func newTestProgram(name string) *entity.Program {
	now := time.Now().UTC()
	status := entity.ProgramStatusActive
	return &entity.Program{
		ID:          uuid.New(),
		Name:        name,
		Description: "Test program description",
		Status:      status,
		StartDate:   &now,
		EndDate:     nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestProgramRepo_Create(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	program := newTestProgram("Integration Test Program")
	err := repo.Create(ctx, program)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify it was persisted
	found, err := repo.GetByID(ctx, program.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected program, got nil")
	}
	if found.Name != program.Name {
		t.Errorf("expected name %q, got %q", program.Name, found.Name)
	}
}

func TestProgramRepo_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for non-existent program")
	}
}

func TestProgramRepo_List(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	// Create 3 programs
	for i := 0; i < 3; i++ {
		if err := repo.Create(ctx, newTestProgram("Program "+string(rune('A'+i)))); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	programs, total, err := repo.List(ctx, repository.ProgramFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(programs) != 3 {
		t.Errorf("expected 3 programs, got %d", len(programs))
	}
}

func TestProgramRepo_List_WithPagination(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := repo.Create(ctx, newTestProgram("Page Program "+string(rune('A'+i)))); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	programs, total, err := repo.List(ctx, repository.ProgramFilter{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(programs) != 2 {
		t.Errorf("expected 2 programs (page 1), got %d", len(programs))
	}
}

func TestProgramRepo_List_WithStatusFilter(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	active := newTestProgram("Active Program")
	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	inactive := newTestProgram("Inactive Program")
	inactiveStatus := entity.ProgramStatusInactive
	inactive.Status = inactiveStatus
	if err := repo.Create(ctx, inactive); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	activeStatus := entity.ProgramStatusActive
	programs, total, err := repo.List(ctx, repository.ProgramFilter{
		Status: &activeStatus,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 active program, got %d", total)
	}
	if len(programs) != 1 {
		t.Errorf("expected 1 program, got %d", len(programs))
	}
}

func TestProgramRepo_Update(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	program := newTestProgram("Original Name")
	if err := repo.Create(ctx, program); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	program.Name = "Updated Name"
	if err := repo.Update(ctx, program); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.GetByID(ctx, program.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.Name != "Updated Name" {
		t.Errorf("expected name %q, got %q", "Updated Name", found.Name)
	}
}

func TestProgramRepo_Delete(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	program := newTestProgram("To Delete")
	if err := repo.Create(ctx, program); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(ctx, program.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should return nil after soft delete
	found, err := repo.GetByID(ctx, program.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil after delete, got program")
	}
}
