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

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	tenantuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/tenant"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTenantHandler(repo *mocks.TenantRepositoryMock, userRepo *mocks.UserRepositoryMock) *handler.TenantHandler {
	createUC := tenantuc.NewCreateTenantUseCase(repo, userRepo)
	getUC := tenantuc.NewGetTenantUseCase(repo)
	listUC := tenantuc.NewListTenantsUseCase(repo)
	updateUC := tenantuc.NewUpdateTenantUseCase(repo)
	deactivateUC := tenantuc.NewDeactivateTenantUseCase(repo)
	deleteUC := tenantuc.NewDeleteTenantUseCase(repo)
	return handler.NewTenantHandler(createUC, getUC, listUC, updateUC, deactivateUC, deleteUC)
}

func withAuthContext(r *http.Request, userID uuid.UUID, roles []string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.AuthUserIDKey, userID.String())
	ctx = context.WithValue(ctx, middleware.AuthRolesKey, roles)
	return r.WithContext(ctx)
}

func TestTenantHandler_Create_Success(t *testing.T) {
	email := "new@example.com"
	landlordID := uuid.New()

	repo := &mocks.TenantRepositoryMock{
		GetByEmailFn: func(ctx context.Context, e string) (*entity.Tenant, error) { return nil, nil },
		CreateFn:     func(ctx context.Context, ten *entity.Tenant) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	h := newTenantHandler(repo, userRepo)
	body := map[string]any{"full_name": "John Doe", "email": email}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader(b))
	req = withAuthContext(req, landlordID, []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTenantHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{}
	h := newTenantHandler(repo, userRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader([]byte("not json")))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTenantHandler_Create_ValidationFailed(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{}
	h := newTenantHandler(repo, userRepo)

	body := map[string]any{"full_name": ""} // missing required full_name
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewReader(b))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTenantHandler_Get_Success(t *testing.T) {
	tenantID := uuid.New()
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{
				ID: tenantID, FullName: "John", Email: &email,
				LandlordID: uuid.New(), IsActive: true,
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}
	h := newTenantHandler(repo, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/tenants/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/"+tenantID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTenantHandler_Get_InvalidID(t *testing.T) {
	h := newTenantHandler(&mocks.TenantRepositoryMock{}, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/tenants/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTenantHandler_Get_NotFound(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}
	h := newTenantHandler(repo, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/tenants/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTenantHandler_List_Success(t *testing.T) {
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
			return []*entity.Tenant{{
				ID: uuid.New(), FullName: "John", Email: &email,
				LandlordID: uuid.New(), IsActive: true,
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			}}, 1, nil
		},
	}
	h := newTenantHandler(repo, &mocks.UserRepositoryMock{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants?limit=10", nil)
	w := httptest.NewRecorder()

	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTenantHandler_Delete_Success(t *testing.T) {
	tenantID := uuid.New()
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, FullName: "John", Email: &email}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	h := newTenantHandler(repo, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Delete("/api/v1/tenants/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/"+tenantID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
