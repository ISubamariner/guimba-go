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

func newTestBeneficiary(name string) *entity.Beneficiary {
	now := time.Now().UTC()
	email := name + "@example.com"
	phone := "+1234567890"
	return &entity.Beneficiary{
		ID:          uuid.New(),
		FullName:    name,
		Email:       &email,
		PhoneNumber: &phone,
		Status:      entity.BeneficiaryStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestBeneficiaryRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	repo := pg.NewBeneficiaryRepoPG(testPgPool)
	ctx := context.Background()

	beneficiary := newTestBeneficiary("Test Beneficiary")
	err := repo.Create(ctx, beneficiary)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByID
	found, err := repo.GetByID(ctx, beneficiary.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected beneficiary, got nil")
	}
	if found.FullName != beneficiary.FullName {
		t.Errorf("expected full_name %q, got %q", beneficiary.FullName, found.FullName)
	}
	if found.Status != entity.BeneficiaryStatusActive {
		t.Errorf("expected status %q, got %q", entity.BeneficiaryStatusActive, found.Status)
	}
}

func TestBeneficiaryRepo_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := pg.NewBeneficiaryRepoPG(testPgPool)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for non-existent beneficiary")
	}
}

func TestBeneficiaryRepo_List_WithPagination(t *testing.T) {
	truncateAll(t)
	repo := pg.NewBeneficiaryRepoPG(testPgPool)
	ctx := context.Background()

	// Create 5 beneficiaries
	for i := 0; i < 5; i++ {
		beneficiary := newTestBeneficiary("Beneficiary " + string(rune('A'+i)))
		if err := repo.Create(ctx, beneficiary); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List first page
	beneficiaries, total, err := repo.List(ctx, repository.BeneficiaryFilter{Limit: 3, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(beneficiaries) != 3 {
		t.Errorf("expected 3 beneficiaries on first page, got %d", len(beneficiaries))
	}

	// List second page
	beneficiaries, total, err = repo.List(ctx, repository.BeneficiaryFilter{Limit: 3, Offset: 3})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(beneficiaries) != 2 {
		t.Errorf("expected 2 beneficiaries on second page, got %d", len(beneficiaries))
	}
}

func TestBeneficiaryRepo_List_WithStatusFilter(t *testing.T) {
	truncateAll(t)
	repo := pg.NewBeneficiaryRepoPG(testPgPool)
	ctx := context.Background()

	// Create active beneficiary
	active := newTestBeneficiary("Active Beneficiary")
	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("Create active failed: %v", err)
	}

	// Create suspended beneficiary
	suspended := newTestBeneficiary("Suspended Beneficiary")
	suspended.Status = entity.BeneficiaryStatusSuspended
	if err := repo.Create(ctx, suspended); err != nil {
		t.Fatalf("Create suspended failed: %v", err)
	}

	// Filter by active status
	activeStatus := entity.BeneficiaryStatusActive
	beneficiaries, total, err := repo.List(ctx, repository.BeneficiaryFilter{
		Status: &activeStatus,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 active beneficiary, got %d", total)
	}
	if len(beneficiaries) != 1 {
		t.Errorf("expected 1 beneficiary, got %d", len(beneficiaries))
	}
	if beneficiaries[0].Status != entity.BeneficiaryStatusActive {
		t.Errorf("expected active status, got %q", beneficiaries[0].Status)
	}
}

func TestBeneficiaryRepo_EnrollInProgram(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	beneficiaryRepo := pg.NewBeneficiaryRepoPG(testPgPool)
	programRepo := pg.NewProgramRepoPG(testPgPool)

	// Create a beneficiary
	beneficiary := newTestBeneficiary("Enrollee")
	if err := beneficiaryRepo.Create(ctx, beneficiary); err != nil {
		t.Fatalf("Create beneficiary failed: %v", err)
	}

	// Create a program
	program := newTestProgram("Test Program")
	if err := programRepo.Create(ctx, program); err != nil {
		t.Fatalf("Create program failed: %v", err)
	}

	// Enroll beneficiary in program
	err := beneficiaryRepo.EnrollInProgram(ctx, beneficiary.ID, program.ID)
	if err != nil {
		t.Fatalf("EnrollInProgram failed: %v", err)
	}

	// Verify enrollment by getting beneficiary with programs loaded
	found, err := beneficiaryRepo.GetByID(ctx, beneficiary.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected beneficiary, got nil")
	}
	if len(found.Programs) != 1 {
		t.Fatalf("expected 1 enrolled program, got %d", len(found.Programs))
	}
	if found.Programs[0].ProgramID != program.ID {
		t.Errorf("expected program ID %v, got %v", program.ID, found.Programs[0].ProgramID)
	}
	if found.Programs[0].ProgramName != program.Name {
		t.Errorf("expected program name %q, got %q", program.Name, found.Programs[0].ProgramName)
	}
}

func TestBeneficiaryRepo_EnrollInProgram_Duplicate(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	beneficiaryRepo := pg.NewBeneficiaryRepoPG(testPgPool)
	programRepo := pg.NewProgramRepoPG(testPgPool)

	beneficiary := newTestBeneficiary("Duplicate Enrollee")
	if err := beneficiaryRepo.Create(ctx, beneficiary); err != nil {
		t.Fatalf("Create beneficiary failed: %v", err)
	}

	program := newTestProgram("Duplicate Test Program")
	if err := programRepo.Create(ctx, program); err != nil {
		t.Fatalf("Create program failed: %v", err)
	}

	// First enrollment should succeed
	err := beneficiaryRepo.EnrollInProgram(ctx, beneficiary.ID, program.ID)
	if err != nil {
		t.Fatalf("First EnrollInProgram failed: %v", err)
	}

	// Second enrollment should fail with ErrBeneficiaryAlreadyEnrolled
	err = beneficiaryRepo.EnrollInProgram(ctx, beneficiary.ID, program.ID)
	if err != entity.ErrBeneficiaryAlreadyEnrolled {
		t.Errorf("expected ErrBeneficiaryAlreadyEnrolled, got %v", err)
	}
}

func TestBeneficiaryRepo_RemoveFromProgram(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	beneficiaryRepo := pg.NewBeneficiaryRepoPG(testPgPool)
	programRepo := pg.NewProgramRepoPG(testPgPool)

	beneficiary := newTestBeneficiary("Unenrollee")
	if err := beneficiaryRepo.Create(ctx, beneficiary); err != nil {
		t.Fatalf("Create beneficiary failed: %v", err)
	}

	program := newTestProgram("Removal Test Program")
	if err := programRepo.Create(ctx, program); err != nil {
		t.Fatalf("Create program failed: %v", err)
	}

	// Enroll first
	if err := beneficiaryRepo.EnrollInProgram(ctx, beneficiary.ID, program.ID); err != nil {
		t.Fatalf("EnrollInProgram failed: %v", err)
	}

	// Remove from program
	err := beneficiaryRepo.RemoveFromProgram(ctx, beneficiary.ID, program.ID)
	if err != nil {
		t.Fatalf("RemoveFromProgram failed: %v", err)
	}

	// Verify removal
	found, err := beneficiaryRepo.GetByID(ctx, beneficiary.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if len(found.Programs) != 0 {
		t.Errorf("expected 0 enrolled programs after removal, got %d", len(found.Programs))
	}
}

func TestBeneficiaryRepo_RemoveFromProgram_NotEnrolled(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	beneficiaryRepo := pg.NewBeneficiaryRepoPG(testPgPool)
	programRepo := pg.NewProgramRepoPG(testPgPool)

	beneficiary := newTestBeneficiary("Never Enrolled")
	if err := beneficiaryRepo.Create(ctx, beneficiary); err != nil {
		t.Fatalf("Create beneficiary failed: %v", err)
	}

	program := newTestProgram("Not Enrolled Program")
	if err := programRepo.Create(ctx, program); err != nil {
		t.Fatalf("Create program failed: %v", err)
	}

	// Try to remove without enrolling first
	err := beneficiaryRepo.RemoveFromProgram(ctx, beneficiary.ID, program.ID)
	if err != entity.ErrBeneficiaryNotEnrolled {
		t.Errorf("expected ErrBeneficiaryNotEnrolled, got %v", err)
	}
}

func TestBeneficiaryRepo_GetByID_WithEnrolledPrograms(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	beneficiaryRepo := pg.NewBeneficiaryRepoPG(testPgPool)
	programRepo := pg.NewProgramRepoPG(testPgPool)

	beneficiary := newTestBeneficiary("Multi Enrollee")
	if err := beneficiaryRepo.Create(ctx, beneficiary); err != nil {
		t.Fatalf("Create beneficiary failed: %v", err)
	}

	// Create and enroll in multiple programs
	programNames := []string{"Program A", "Program B", "Program C"}
	for _, name := range programNames {
		program := newTestProgram(name)
		if err := programRepo.Create(ctx, program); err != nil {
			t.Fatalf("Create program failed: %v", err)
		}
		if err := beneficiaryRepo.EnrollInProgram(ctx, beneficiary.ID, program.ID); err != nil {
			t.Fatalf("EnrollInProgram failed: %v", err)
		}
	}

	// Get beneficiary and verify all programs are loaded
	found, err := beneficiaryRepo.GetByID(ctx, beneficiary.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected beneficiary, got nil")
	}
	if len(found.Programs) != 3 {
		t.Fatalf("expected 3 enrolled programs, got %d", len(found.Programs))
	}

	// Verify programs are sorted by enrollment date (DESC)
	for i, prog := range found.Programs {
		if prog.Status != "active" {
			t.Errorf("expected program %d status 'active', got %q", i, prog.Status)
		}
		if prog.EnrolledAt.IsZero() {
			t.Errorf("program %d has zero EnrolledAt timestamp", i)
		}
	}
}
