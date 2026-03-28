package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	transaction "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- RecordPayment ---

func TestRecordPayment_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	d := newTestDebt(tenantID, landlordID, 1000)

	txRepo := &mocks.TransactionRepositoryMock{
		CreateFn: func(ctx context.Context, tx *entity.Transaction) error { return nil },
		ExistsByReferenceNumberFn: func(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
			return false, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})

	amount, _ := entity.NewMoney(decimal.NewFromFloat(300), entity.CurrencyPHP)
	recorderID := uuid.New()
	tx, err := uc.Execute(context.Background(), d.ID, tenantID, &recorderID, amount, entity.PaymentMethodCash, time.Now(), "Partial payment", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx == nil {
		t.Fatal("expected transaction to be returned")
	}
	if d.Status != entity.DebtStatusPartial {
		t.Errorf("expected debt status PARTIAL, got %s", d.Status)
	}
}

func TestRecordPayment_DebtNotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Payment", nil, nil)
	if err == nil {
		t.Fatal("expected error when debt not found")
	}
}

func TestRecordPayment_Overpayment(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 100)

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	amount, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Over", nil, nil)
	if err != entity.ErrDebtOverpayment {
		t.Fatalf("expected ErrDebtOverpayment, got %v", err)
	}
}

func TestRecordPayment_DuplicateReference(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 1000)

	txRepo := &mocks.TransactionRepositoryMock{
		ExistsByReferenceNumberFn: func(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
			return true, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	ref := "REF-001"
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Payment", nil, &ref)
	if err != entity.ErrTransactionDuplicateReference {
		t.Fatalf("expected ErrTransactionDuplicateReference, got %v", err)
	}
}

func TestRecordPayment_AlreadyPaid(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 100)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = d.RecordPayment(payment) // PAID

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	amount, _ := entity.NewMoney(decimal.NewFromFloat(50), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Late", nil, nil)
	if err != entity.ErrDebtAlreadyPaid {
		t.Fatalf("expected ErrDebtAlreadyPaid, got %v", err)
	}
}

// --- RecordRefund ---

func TestRecordRefund_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	d := newTestDebt(tenantID, landlordID, 1000)
	// Pre-pay 500
	payment, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	_ = d.RecordPayment(payment)

	txRepo := &mocks.TransactionRepositoryMock{
		CreateFn: func(ctx context.Context, tx *entity.Transaction) error { return nil },
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}

	uc := transaction.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	refundAmount, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	recorderID := uuid.New()
	tx, err := uc.Execute(context.Background(), d.ID, tenantID, &recorderID, refundAmount, entity.PaymentMethodCash, time.Now(), "Partial refund", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx == nil {
		t.Fatal("expected transaction returned")
	}
	if d.Status != entity.DebtStatusPartial {
		t.Errorf("expected PARTIAL after partial refund of partial payment, got %s", d.Status)
	}
}

func TestRecordRefund_ZeroPayments(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 1000) // no payments made

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	refundAmount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, refundAmount, entity.PaymentMethodCash, time.Now(), "Refund", nil)
	if err != entity.ErrInsufficientAmount {
		t.Fatalf("expected ErrInsufficientAmount, got %v", err)
	}
}

func TestRecordRefund_ExceedsAmountPaid(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 1000)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = d.RecordPayment(payment)

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo, &mocks.AuditRepositoryMock{})
	refundAmount, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, refundAmount, entity.PaymentMethodCash, time.Now(), "Too much", nil)
	if err != entity.ErrInsufficientAmount {
		t.Fatalf("expected ErrInsufficientAmount, got %v", err)
	}
}

// --- GetTransaction ---

func TestGetTransaction_Success(t *testing.T) {
	txID := uuid.New()
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	expected, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)
	expected.ID = txID

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return expected, nil
		},
	}
	uc := transaction.NewGetTransactionUseCase(txRepo)
	result, err := uc.Execute(context.Background(), txID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != txID {
		t.Error("expected ID to match")
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return nil, nil
		},
	}
	uc := transaction.NewGetTransactionUseCase(txRepo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- ListTransactions ---

func TestListTransactions_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx1, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
			return []*entity.Transaction{tx1}, 1, nil
		},
	}
	uc := transaction.NewListTransactionsUseCase(txRepo)
	txs, total, err := uc.Execute(context.Background(), repository.TransactionFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 || len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}

func TestListTransactions_DefaultLimit(t *testing.T) {
	var captured repository.TransactionFilter
	txRepo := &mocks.TransactionRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := transaction.NewListTransactionsUseCase(txRepo)
	_, _, _ = uc.Execute(context.Background(), repository.TransactionFilter{Limit: 0})
	if captured.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", captured.Limit)
	}
}

// --- VerifyTransaction ---

func TestVerifyTransaction_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return tx, nil
		},
		UpdateFn: func(ctx context.Context, t *entity.Transaction) error { return nil },
	}
	uc := transaction.NewVerifyTransactionUseCase(txRepo, &mocks.AuditRepositoryMock{})
	verifierID := uuid.New()
	err := uc.Execute(context.Background(), tx.ID, verifierID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tx.IsVerified {
		t.Error("expected transaction to be verified")
	}
}

func TestVerifyTransaction_AlreadyVerified(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)
	_ = tx.Verify(uuid.New())

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return tx, nil
		},
	}
	uc := transaction.NewVerifyTransactionUseCase(txRepo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), tx.ID, uuid.New())
	if err != entity.ErrTransactionAlreadyVerified {
		t.Fatalf("expected ErrTransactionAlreadyVerified, got %v", err)
	}
}

func TestVerifyTransaction_NotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return nil, nil
		},
	}
	uc := transaction.NewVerifyTransactionUseCase(txRepo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}
