package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	txuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTransactionHandler(txRepo *mocks.TransactionRepositoryMock, debtRepo *mocks.DebtRepositoryMock, userRepo *mocks.UserRepositoryMock, tenantRepo *mocks.TenantRepositoryMock) *handler.TransactionHandler {
	recordPaymentUC := txuc.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	recordRefundUC := txuc.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	getUC := txuc.NewGetTransactionUseCase(txRepo)
	listUC := txuc.NewListTransactionsUseCase(txRepo)
	verifyUC := txuc.NewVerifyTransactionUseCase(txRepo)
	return handler.NewTransactionHandler(recordPaymentUC, recordRefundUC, getUC, listUC, verifyUC)
}

func TestTransactionHandler_RecordPayment_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	debtID := uuid.New()

	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	d, _ := entity.NewDebt(tenantID, landlordID, nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(30*24*time.Hour), nil)
	d.ID = debtID

	txRepo := &mocks.TransactionRepositoryMock{
		CreateFn: func(ctx context.Context, tx *entity.Transaction) error { return nil },
		ExistsByReferenceNumberFn: func(ctx context.Context, dID uuid.UUID, ref string) (bool, error) {
			return false, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return d, nil },
		UpdateFn:  func(ctx context.Context, debt *entity.Debt) error { return nil },
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

	h := newTransactionHandler(txRepo, debtRepo, userRepo, tenantRepo)
	body := map[string]any{
		"debt_id":          debtID.String(),
		"tenant_id":        tenantID.String(),
		"amount":           map[string]any{"amount": "300.00", "currency": "PHP"},
		"payment_method":   "CASH",
		"transaction_date": time.Now().Format("2006-01-02"),
		"description":      "Partial payment",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/payment", bytes.NewReader(b))
	req = withAuthContext(req, landlordID, []string{"landlord"})
	w := httptest.NewRecorder()
	h.RecordPayment(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_RecordPayment_InvalidJSON(t *testing.T) {
	h := newTransactionHandler(&mocks.TransactionRepositoryMock{}, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/payment", bytes.NewReader([]byte("bad")))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.RecordPayment(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTransactionHandler_Get_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return tx, nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/transactions/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/"+tx.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_Get_NotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return nil, nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/transactions/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTransactionHandler_List_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
			return []*entity.Transaction{tx}, 1, nil
		},
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?limit=10", nil)
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_Verify_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return tx, nil },
		UpdateFn:  func(ctx context.Context, t *entity.Transaction) error { return nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Put("/api/v1/transactions/{id}/verify", h.Verify)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+tx.ID.String()+"/verify", nil)
	req = withAuthContext(req, uuid.New(), []string{"admin"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_Verify_AlreadyVerified(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)
	_ = tx.Verify(uuid.New())

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return tx, nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Put("/api/v1/transactions/{id}/verify", h.Verify)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+tx.ID.String()+"/verify", nil)
	req = withAuthContext(req, uuid.New(), []string{"admin"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}
