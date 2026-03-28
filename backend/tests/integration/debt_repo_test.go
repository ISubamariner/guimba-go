//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
)

func newTestDebt(tenantID, landlordID uuid.UUID) *entity.Debt {
	now := time.Now().UTC()
	amount, _ := entity.NewMoney(decimal.NewFromInt(1000), entity.CurrencyPHP)
	return &entity.Debt{
		ID:             uuid.New(),
		TenantID:       tenantID,
		LandlordID:     landlordID,
		PropertyID:     nil,
		DebtType:       entity.DebtTypeRent,
		Description:    "Monthly rent payment",
		OriginalAmount: amount,
		AmountPaid:     entity.ZeroMoney(entity.CurrencyPHP),
		DueDate:        now.AddDate(0, 0, 30),
		Status:         entity.DebtStatusPending,
		Notes:          nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func TestDebtRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlordID := createTestLandlord(t)

	// Create tenant
	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	tenant := newTestTenant("Test Tenant", landlordID)
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Create debt
	debtRepo := pg.NewDebtRepoPG(testPgPool)
	debt := newTestDebt(tenant.ID, landlordID)
	err := debtRepo.Create(ctx, debt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByID
	found, err := debtRepo.GetByID(ctx, debt.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected debt, got nil")
	}
	if found.TenantID != tenant.ID {
		t.Errorf("expected tenant_id %v, got %v", tenant.ID, found.TenantID)
	}
	if found.LandlordID != landlordID {
		t.Errorf("expected landlord_id %v, got %v", landlordID, found.LandlordID)
	}
	if found.Description != debt.Description {
		t.Errorf("expected description %q, got %q", debt.Description, found.Description)
	}
	if found.Status != entity.DebtStatusPending {
		t.Errorf("expected status PENDING, got %v", found.Status)
	}
	if found.OriginalAmount.Amount.Cmp(debt.OriginalAmount.Amount) != 0 {
		t.Errorf("expected amount %v, got %v", debt.OriginalAmount.Amount, found.OriginalAmount.Amount)
	}
	if found.OriginalAmount.Currency != entity.CurrencyPHP {
		t.Errorf("expected currency PHP, got %v", found.OriginalAmount.Currency)
	}
}

func TestDebtRepo_List_WithTenantFilter(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlordID := createTestLandlord(t)
	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	debtRepo := pg.NewDebtRepoPG(testPgPool)

	// Create 2 tenants
	tenant1 := newTestTenant("Tenant One", landlordID)
	if err := tenantRepo.Create(ctx, tenant1); err != nil {
		t.Fatalf("Failed to create tenant1: %v", err)
	}
	tenant2 := newTestTenant("Tenant Two", landlordID)
	if err := tenantRepo.Create(ctx, tenant2); err != nil {
		t.Fatalf("Failed to create tenant2: %v", err)
	}

	// Create 2 debts for tenant1
	for i := 0; i < 2; i++ {
		debt := newTestDebt(tenant1.ID, landlordID)
		if err := debtRepo.Create(ctx, debt); err != nil {
			t.Fatalf("Create debt failed: %v", err)
		}
	}

	// Create 1 debt for tenant2
	debt := newTestDebt(tenant2.ID, landlordID)
	if err := debtRepo.Create(ctx, debt); err != nil {
		t.Fatalf("Create debt failed: %v", err)
	}

	// List debts for tenant1
	debts, total, err := debtRepo.List(ctx, repository.DebtFilter{
		TenantID: &tenant1.ID,
		Limit:    10,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 debts for tenant1, got %d", total)
	}
	if len(debts) != 2 {
		t.Errorf("expected 2 debts, got %d", len(debts))
	}

	// List debts for tenant2
	debts, total, err = debtRepo.List(ctx, repository.DebtFilter{
		TenantID: &tenant2.ID,
		Limit:    10,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 debt for tenant2, got %d", total)
	}
	if len(debts) != 1 {
		t.Errorf("expected 1 debt, got %d", len(debts))
	}
}

func TestDebtRepo_Update(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlordID := createTestLandlord(t)
	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	tenant := newTestTenant("Test Tenant", landlordID)
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Create debt
	debtRepo := pg.NewDebtRepoPG(testPgPool)
	debt := newTestDebt(tenant.ID, landlordID)
	if err := debtRepo.Create(ctx, debt); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the debt - simulate payment
	payment, _ := entity.NewMoney(decimal.NewFromInt(500), entity.CurrencyPHP)
	err := debt.RecordPayment(payment)
	if err != nil {
		t.Fatalf("RecordPayment failed: %v", err)
	}

	// Persist the update
	if err := debtRepo.Update(ctx, debt); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify the update persisted
	found, err := debtRepo.GetByID(ctx, debt.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.Status != entity.DebtStatusPartial {
		t.Errorf("expected status PARTIAL, got %v", found.Status)
	}
	if found.AmountPaid.Amount.Cmp(decimal.NewFromInt(500)) != 0 {
		t.Errorf("expected amount_paid 500, got %v", found.AmountPaid.Amount)
	}
}

func TestDebtRepo_Delete(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlordID := createTestLandlord(t)
	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	tenant := newTestTenant("Test Tenant", landlordID)
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Create debt
	debtRepo := pg.NewDebtRepoPG(testPgPool)
	debt := newTestDebt(tenant.ID, landlordID)
	if err := debtRepo.Create(ctx, debt); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the debt (soft delete)
	if err := debtRepo.Delete(ctx, debt.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify GetByID returns nil
	found, err := debtRepo.GetByID(ctx, debt.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil after delete, got %+v", found)
	}
}
