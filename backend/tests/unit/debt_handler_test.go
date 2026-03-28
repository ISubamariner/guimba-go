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
	debtuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newDebtHandler(debtRepo *mocks.DebtRepositoryMock, userRepo *mocks.UserRepositoryMock, tenantRepo *mocks.TenantRepositoryMock, propRepo *mocks.PropertyRepositoryMock) *handler.DebtHandler {
	auditRepo := &mocks.AuditRepositoryMock{}
	createUC := debtuc.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo, auditRepo)
	getUC := debtuc.NewGetDebtUseCase(debtRepo)
	listUC := debtuc.NewListDebtsUseCase(debtRepo)
	updateUC := debtuc.NewUpdateDebtUseCase(debtRepo, auditRepo)
	cancelUC := debtuc.NewCancelDebtUseCase(debtRepo, auditRepo)
	markPaidUC := debtuc.NewMarkDebtPaidUseCase(debtRepo, auditRepo)
	deleteUC := debtuc.NewDeleteDebtUseCase(debtRepo, auditRepo)
	return handler.NewDebtHandler(createUC, getUC, listUC, updateUC, cancelUC, markPaidUC, deleteUC)
}

func newTestDebtEntity() *entity.Debt {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	d, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Monthly rent", amount, time.Now().Add(30*24*time.Hour), nil)
	return d
}

func TestDebtHandler_Create_Success(t *testing.T) {
	landlordID := uuid.New()
	tenantID := uuid.New()

	debtRepo := &mocks.DebtRepositoryMock{
		CreateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: landlordID, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{}

	h := newDebtHandler(debtRepo, userRepo, tenantRepo, propRepo)
	body := map[string]any{
		"tenant_id":       tenantID.String(),
		"debt_type":       "RENT",
		"description":     "Monthly rent",
		"original_amount": map[string]any{"amount": "1000.00", "currency": "PHP"},
		"due_date":        time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02"),
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewReader(b))
	req = withAuthContext(req, landlordID, []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Create_InvalidJSON(t *testing.T) {
	h := newDebtHandler(&mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewReader([]byte("bad")))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDebtHandler_Create_ValidationFailed(t *testing.T) {
	h := newDebtHandler(&mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	body := map[string]any{"description": ""} // missing required fields
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewReader(b))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Get_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/debts/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts/"+d.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Get_InvalidID(t *testing.T) {
	h := newDebtHandler(&mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	r := chi.NewRouter()
	r.Get("/api/v1/debts/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts/not-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDebtHandler_Get_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return nil, nil },
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	r := chi.NewRouter()
	r.Get("/api/v1/debts/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDebtHandler_List_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			return []*entity.Debt{d}, 1, nil
		},
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts?limit=10", nil)
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Cancel_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return d, nil },
		UpdateFn:  func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})

	body := map[string]any{"reason": "No longer valid"}
	b, _ := json.Marshal(body)

	r := chi.NewRouter()
	r.Put("/api/v1/debts/{id}/cancel", h.Cancel)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/debts/"+d.ID.String()+"/cancel", bytes.NewReader(b))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Delete_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return d, nil },
		DeleteFn:  func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})

	r := chi.NewRouter()
	r.Delete("/api/v1/debts/{id}", h.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/debts/"+d.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
