package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	debt "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	property "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- helpers ---

func newTestDebt(tenantID, landlordID uuid.UUID, amount float64) *entity.Debt {
	m, _ := entity.NewMoney(decimal.NewFromFloat(amount), entity.CurrencyPHP)
	d, _ := entity.NewDebt(tenantID, landlordID, nil, entity.DebtTypeRent, "Test rent", m, time.Now().Add(30*24*time.Hour), nil)
	return d
}

// --- CreateDebt ---

func TestCreateDebt_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()

	debtRepo := &mocks.DebtRepositoryMock{
		CreateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: landlordID, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{}

	uc := debt.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	d := newTestDebt(tenantID, landlordID, 1000)
	err := uc.Execute(context.Background(), d)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateDebt_TenantNotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{}

	uc := debt.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	d := newTestDebt(uuid.New(), uuid.New(), 1000)
	err := uc.Execute(context.Background(), d)
	if err == nil {
		t.Fatal("expected error when tenant not found")
	}
}

func TestCreateDebt_WithProperty_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	propID := uuid.New()

	debtRepo := &mocks.DebtRepositoryMock{
		CreateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: landlordID, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, OwnerID: landlordID, IsActive: true}, nil
		},
	}

	uc := debt.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	m, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	d, _ := entity.NewDebt(tenantID, landlordID, &propID, entity.DebtTypeRent, "Rent", m, time.Now().Add(30*24*time.Hour), nil)
	err := uc.Execute(context.Background(), d)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- GetDebt ---

func TestGetDebt_Success(t *testing.T) {
	debtID := uuid.New()
	d := newTestDebt(uuid.New(), uuid.New(), 500)
	d.ID = debtID

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	uc := debt.NewGetDebtUseCase(debtRepo)
	result, err := uc.Execute(context.Background(), debtID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != debtID {
		t.Error("expected ID to match")
	}
}

func TestGetDebt_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	uc := debt.NewGetDebtUseCase(debtRepo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestGetDebt_LazyOverdue(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 500)
	d.DueDate = time.Now().Add(-48 * time.Hour)
	d.Status = entity.DebtStatusPending

	var updatedDebt *entity.Debt
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error {
			updatedDebt = debt
			return nil
		},
	}
	uc := debt.NewGetDebtUseCase(debtRepo)
	result, err := uc.Execute(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != entity.DebtStatusOverdue {
		t.Errorf("expected OVERDUE, got %s", result.Status)
	}
	if updatedDebt == nil {
		t.Error("expected update to be called for lazy overdue")
	}
}

// --- ListDebts ---

func TestListDebts_Success(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			d := newTestDebt(uuid.New(), uuid.New(), 100)
			return []*entity.Debt{d}, 1, nil
		},
	}
	uc := debt.NewListDebtsUseCase(debtRepo)
	debts, total, err := uc.Execute(context.Background(), repository.DebtFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 || len(debts) != 1 {
		t.Errorf("expected 1 debt, got %d", len(debts))
	}
}

func TestListDebts_DefaultLimit(t *testing.T) {
	var captured repository.DebtFilter
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := debt.NewListDebtsUseCase(debtRepo)
	_, _, _ = uc.Execute(context.Background(), repository.DebtFilter{Limit: 0})
	if captured.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", captured.Limit)
	}
}

func TestListDebts_MaxLimit(t *testing.T) {
	var captured repository.DebtFilter
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := debt.NewListDebtsUseCase(debtRepo)
	_, _, _ = uc.Execute(context.Background(), repository.DebtFilter{Limit: 999})
	if captured.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", captured.Limit)
	}
}

// --- UpdateDebt ---

func TestUpdateDebt_Success(t *testing.T) {
	debtID := uuid.New()
	existing := newTestDebt(uuid.New(), uuid.New(), 1000)
	existing.ID = debtID

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	uc := debt.NewUpdateDebtUseCase(debtRepo)

	updates := &entity.Debt{Description: "Updated rent description", DebtType: entity.DebtTypeRent, DueDate: time.Now().Add(60 * 24 * time.Hour)}
	err := uc.Execute(context.Background(), debtID, updates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateDebt_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	uc := debt.NewUpdateDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), uuid.New(), &entity.Debt{Description: "x", DebtType: entity.DebtTypeRent, DueDate: time.Now().Add(24 * time.Hour)})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- CancelDebt ---

func TestCancelDebt_Success(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 500)
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error {
			if d.Status != entity.DebtStatusCancelled {
				t.Errorf("expected CANCELLED, got %s", d.Status)
			}
			return nil
		},
	}
	uc := debt.NewCancelDebtUseCase(debtRepo)
	reason := "No longer valid"
	err := uc.Execute(context.Background(), existing.ID, &reason)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCancelDebt_AlreadyPaid(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 100)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = existing.RecordPayment(payment) // PAID

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
	}
	uc := debt.NewCancelDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID, nil)
	if err != entity.ErrDebtAlreadyPaid {
		t.Fatalf("expected ErrDebtAlreadyPaid, got %v", err)
	}
}

func TestCancelDebt_AlreadyCancelled(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 100)
	reason := "first cancel"
	_ = existing.Cancel(&reason) // CANCELLED

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	uc := debt.NewCancelDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID, nil)
	// Cancel on already-cancelled is idempotent (no error), status stays CANCELLED
	if err != nil {
		t.Fatalf("expected no error on re-cancel, got %v", err)
	}
}

// --- MarkDebtPaid ---

func TestMarkDebtPaid_Success(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 500)
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error {
			if d.Status != entity.DebtStatusPaid {
				t.Errorf("expected PAID, got %s", d.Status)
			}
			return nil
		},
	}
	uc := debt.NewMarkDebtPaidUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- DeleteDebt ---

func TestDeleteDebt_Success(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 100)
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	uc := debt.NewDeleteDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteDebt_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	uc := debt.NewDeleteDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- Property Deactivation with Debt Check ---

func TestDeactivateProperty_BlockedByActiveDebts(t *testing.T) {
	propID := uuid.New()
	propRepo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, IsActive: true}, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		HasActiveDebtsForPropertyFn: func(ctx context.Context, propertyID uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(propRepo, debtRepo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), propID)
	if err == nil {
		t.Fatal("expected error when property has active debts")
	}
}

func TestDeactivateProperty_AllowedWhenNoActiveDebts(t *testing.T) {
	propID := uuid.New()
	propRepo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, IsActive: true}, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Property) error { return nil },
	}
	debtRepo := &mocks.DebtRepositoryMock{
		HasActiveDebtsForPropertyFn: func(ctx context.Context, propertyID uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(propRepo, debtRepo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
