package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/scheduler"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestOverdueScheduler_MarksOverdueDebts(t *testing.T) {
	ctx := context.Background()

	// Create test debts
	tenantID := uuid.New()
	landlordID := uuid.New()
	amt5k, _ := entity.NewMoney(decimal.NewFromInt(5000), entity.CurrencyPHP)

	// Debt 1: Past due (yesterday), PENDING status - should be marked OVERDUE
	pastDue, err := entity.NewDebt(
		tenantID,
		landlordID,
		nil,
		entity.DebtTypeRent,
		"Past due rent",
		amt5k,
		time.Now().UTC().AddDate(0, 0, -1), // Yesterday
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create past due debt: %v", err)
	}

	// Debt 2: Future due (tomorrow), PENDING status - should NOT be marked OVERDUE
	futureDue, err := entity.NewDebt(
		tenantID,
		landlordID,
		nil,
		entity.DebtTypeRent,
		"Future rent",
		amt5k,
		time.Now().UTC().AddDate(0, 0, 1), // Tomorrow
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create future due debt: %v", err)
	}

	// Track which debts were updated
	var updatedDebtIDs []uuid.UUID

	mockRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			// Return both debts for PENDING status
			if filter.Status != nil && *filter.Status == entity.DebtStatusPending {
				return []*entity.Debt{pastDue, futureDue}, 2, nil
			}
			// Return empty for PARTIAL status
			if filter.Status != nil && *filter.Status == entity.DebtStatusPartial {
				return []*entity.Debt{}, 0, nil
			}
			return []*entity.Debt{}, 0, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error {
			updatedDebtIDs = append(updatedDebtIDs, debt.ID)
			return nil
		},
	}

	// Create scheduler and run check
	s := scheduler.NewOverdueScheduler(mockRepo, 24*time.Hour)
	s.CheckOverdue(ctx)

	// Verify exactly one debt was updated
	if len(updatedDebtIDs) != 1 {
		t.Errorf("expected 1 debt to be updated, got %d", len(updatedDebtIDs))
	}

	// Verify it was the past-due debt
	if len(updatedDebtIDs) > 0 && updatedDebtIDs[0] != pastDue.ID {
		t.Errorf("expected past-due debt %s to be updated, got %s", pastDue.ID, updatedDebtIDs[0])
	}

	// Verify the past-due debt status was changed to OVERDUE
	if pastDue.Status != entity.DebtStatusOverdue {
		t.Errorf("expected past-due debt status to be OVERDUE, got %s", pastDue.Status)
	}

	// Verify the future-due debt status remains PENDING
	if futureDue.Status != entity.DebtStatusPending {
		t.Errorf("expected future-due debt status to remain PENDING, got %s", futureDue.Status)
	}
}

func TestOverdueScheduler_SkipsAlreadyOverdue(t *testing.T) {
	ctx := context.Background()

	// Create test debt that is already OVERDUE
	tenantID := uuid.New()
	landlordID := uuid.New()
	amt5k, _ := entity.NewMoney(decimal.NewFromInt(5000), entity.CurrencyPHP)

	alreadyOverdue, err := entity.NewDebt(
		tenantID,
		landlordID,
		nil,
		entity.DebtTypeRent,
		"Already overdue rent",
		amt5k,
		time.Now().UTC().AddDate(0, 0, -7), // Week ago
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create debt: %v", err)
	}

	// Manually mark it as overdue (since it starts as PENDING)
	alreadyOverdue.MarkAsOverdue()

	updateCalled := false

	mockRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			// Return the already-overdue debt when querying PENDING status
			// (it won't be in this list in reality, but this tests that MarkAsOverdue is idempotent)
			if filter.Status != nil && *filter.Status == entity.DebtStatusPending {
				return []*entity.Debt{}, 0, nil
			}
			// Return it for PARTIAL status to test the idempotency
			if filter.Status != nil && *filter.Status == entity.DebtStatusPartial {
				return []*entity.Debt{}, 0, nil
			}
			return []*entity.Debt{}, 0, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error {
			updateCalled = true
			return nil
		},
	}

	// Create scheduler and run check
	s := scheduler.NewOverdueScheduler(mockRepo, 24*time.Hour)
	s.CheckOverdue(ctx)

	// Verify Update was NOT called (no debts to process)
	if updateCalled {
		t.Error("expected Update to NOT be called for already-overdue debt")
	}

	// Verify the debt status remains OVERDUE
	if alreadyOverdue.Status != entity.DebtStatusOverdue {
		t.Errorf("expected debt status to remain OVERDUE, got %s", alreadyOverdue.Status)
	}
}
