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

// createTestDebt creates a debt with all required FK dependencies (landlord, tenant).
func createTestDebt(t *testing.T, tenantID, landlordID uuid.UUID) *entity.Debt {
	t.Helper()
	ctx := context.Background()
	debtRepo := pg.NewDebtRepoPG(testPgPool)

	debt := newTestDebt(tenantID, landlordID)
	if err := debtRepo.Create(ctx, debt); err != nil {
		t.Fatalf("Failed to create debt: %v", err)
	}
	return debt
}

func newTestTransaction(debtID, tenantID, landlordID uuid.UUID) *entity.Transaction {
	now := time.Now().UTC()
	amount, _ := entity.NewMoney(decimal.NewFromInt(500), entity.CurrencyPHP)
	return &entity.Transaction{
		ID:               uuid.New(),
		DebtID:           debtID,
		TenantID:         tenantID,
		LandlordID:       landlordID,
		RecordedByUserID: nil,
		TransactionType:  entity.TransactionTypePayment,
		Amount:           amount,
		PaymentMethod:    entity.PaymentMethodCash,
		TransactionDate:  now,
		Description:      "Rent payment",
		ReceiptNumber:    nil,
		ReferenceNumber:  nil,
		IsVerified:       false,
		VerifiedByUserID: nil,
		VerifiedAt:       nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestTransactionRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	// Create prerequisites: landlord -> tenant -> debt
	landlordID := createTestLandlord(t)
	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	tenant := newTestTenant("Test Tenant", landlordID)
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}
	debt := createTestDebt(t, tenant.ID, landlordID)

	// Create transaction
	txRepo := pg.NewTransactionRepoPG(testPgPool)
	tx := newTestTransaction(debt.ID, tenant.ID, landlordID)
	err := txRepo.Create(ctx, tx)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify GetByID
	found, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected transaction, got nil")
	}
	if found.DebtID != debt.ID {
		t.Errorf("expected debt_id %v, got %v", debt.ID, found.DebtID)
	}
	if found.TenantID != tenant.ID {
		t.Errorf("expected tenant_id %v, got %v", tenant.ID, found.TenantID)
	}
	if found.LandlordID != landlordID {
		t.Errorf("expected landlord_id %v, got %v", landlordID, found.LandlordID)
	}
	if found.TransactionType != entity.TransactionTypePayment {
		t.Errorf("expected transaction_type PAYMENT, got %v", found.TransactionType)
	}
	if found.Amount.Amount.Cmp(decimal.NewFromInt(500)) != 0 {
		t.Errorf("expected amount 500, got %v", found.Amount.Amount)
	}
	if found.PaymentMethod != entity.PaymentMethodCash {
		t.Errorf("expected payment_method CASH, got %v", found.PaymentMethod)
	}
	if found.IsVerified {
		t.Error("expected is_verified false, got true")
	}
}

func TestTransactionRepo_List_ByLandlord(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	landlord1 := createTestLandlord(t)
	landlord2 := createTestLandlord(t)

	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	txRepo := pg.NewTransactionRepoPG(testPgPool)

	// Create tenant and debt for landlord1
	tenant1 := newTestTenant("Landlord1 Tenant", landlord1)
	if err := tenantRepo.Create(ctx, tenant1); err != nil {
		t.Fatalf("Failed to create tenant1: %v", err)
	}
	debt1 := createTestDebt(t, tenant1.ID, landlord1)

	// Create 2 transactions for landlord1
	for i := 0; i < 2; i++ {
		tx := newTestTransaction(debt1.ID, tenant1.ID, landlord1)
		if err := txRepo.Create(ctx, tx); err != nil {
			t.Fatalf("Create transaction failed: %v", err)
		}
	}

	// Create tenant and debt for landlord2
	tenant2 := newTestTenant("Landlord2 Tenant", landlord2)
	if err := tenantRepo.Create(ctx, tenant2); err != nil {
		t.Fatalf("Failed to create tenant2: %v", err)
	}
	debt2 := createTestDebt(t, tenant2.ID, landlord2)

	// Create 1 transaction for landlord2
	tx := newTestTransaction(debt2.ID, tenant2.ID, landlord2)
	if err := txRepo.Create(ctx, tx); err != nil {
		t.Fatalf("Create transaction failed: %v", err)
	}

	// List transactions for landlord1
	txs, total, err := txRepo.List(ctx, repository.TransactionFilter{
		LandlordID: &landlord1,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 transactions for landlord1, got %d", total)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}

	// List transactions for landlord2
	txs, total, err = txRepo.List(ctx, repository.TransactionFilter{
		LandlordID: &landlord2,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 transaction for landlord2, got %d", total)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}

func TestTransactionRepo_Verify(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()

	// Create prerequisites
	landlordID := createTestLandlord(t)
	tenantRepo := pg.NewTenantRepoPG(testPgPool)
	tenant := newTestTenant("Test Tenant", landlordID)
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}
	debt := createTestDebt(t, tenant.ID, landlordID)

	// Create transaction
	txRepo := pg.NewTransactionRepoPG(testPgPool)
	tx := newTestTransaction(debt.ID, tenant.ID, landlordID)
	if err := txRepo.Create(ctx, tx); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify the transaction (using domain method)
	verifierID := uuid.New()
	if err := tx.Verify(verifierID); err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	// Persist the update
	if err := txRepo.Update(ctx, tx); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify the update persisted
	found, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !found.IsVerified {
		t.Error("expected is_verified true, got false")
	}
	if found.VerifiedByUserID == nil {
		t.Error("expected verified_by_user_id to be set, got nil")
	} else if *found.VerifiedByUserID != verifierID {
		t.Errorf("expected verified_by_user_id %v, got %v", verifierID, *found.VerifiedByUserID)
	}
	if found.VerifiedAt == nil {
		t.Error("expected verified_at to be set, got nil")
	}
}
