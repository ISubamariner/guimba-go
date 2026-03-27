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
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	propertyuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newPropertyHandler(repo *mocks.PropertyRepositoryMock, userRepo *mocks.UserRepositoryMock) *handler.PropertyHandler {
	createUC := propertyuc.NewCreatePropertyUseCase(repo, userRepo)
	getUC := propertyuc.NewGetPropertyUseCase(repo)
	listUC := propertyuc.NewListPropertiesUseCase(repo)
	updateUC := propertyuc.NewUpdatePropertyUseCase(repo)
	deactivateUC := propertyuc.NewDeactivatePropertyUseCase(repo)
	deleteUC := propertyuc.NewDeletePropertyUseCase(repo)
	return handler.NewPropertyHandler(createUC, getUC, listUC, updateUC, deactivateUC, deleteUC)
}

func TestPropertyHandler_Create_Success(t *testing.T) {
	ownerID := uuid.New()

	repo := &mocks.PropertyRepositoryMock{
		GetByPropertyCodeFn: func(ctx context.Context, code string) (*entity.Property, error) { return nil, nil },
		CreateFn:            func(ctx context.Context, p *entity.Property) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	h := newPropertyHandler(repo, userRepo)
	body := map[string]any{"name": "Farm Plot A", "property_code": "FP-001", "size_in_sqm": 500.0}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/properties", bytes.NewReader(b))
	req = withAuthContext(req, ownerID, []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPropertyHandler_Create_InvalidJSON(t *testing.T) {
	h := newPropertyHandler(&mocks.PropertyRepositoryMock{}, &mocks.UserRepositoryMock{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/properties", bytes.NewReader([]byte("not json")))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPropertyHandler_Create_ValidationFailed(t *testing.T) {
	h := newPropertyHandler(&mocks.PropertyRepositoryMock{}, &mocks.UserRepositoryMock{})

	body := map[string]any{"name": ""} // missing required fields
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/properties", bytes.NewReader(b))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPropertyHandler_Get_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{
				ID: propID, Name: "Farm", PropertyCode: "FP-001",
				SizeInSqm: 500, PropertyType: "LAND", OwnerID: uuid.New(),
				IsActive: true, IsAvailableForRent: true,
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}
	h := newPropertyHandler(repo, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/properties/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties/"+propID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPropertyHandler_Get_InvalidID(t *testing.T) {
	h := newPropertyHandler(&mocks.PropertyRepositoryMock{}, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/properties/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPropertyHandler_Get_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	h := newPropertyHandler(repo, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/properties/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestPropertyHandler_List_Success(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			return []*entity.Property{{
				ID: uuid.New(), Name: "Farm", PropertyCode: "FP-001",
				SizeInSqm: 500, PropertyType: "LAND", OwnerID: uuid.New(),
				IsActive: true, IsAvailableForRent: true,
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			}}, 1, nil
		},
	}
	h := newPropertyHandler(repo, &mocks.UserRepositoryMock{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties?limit=10", nil)
	w := httptest.NewRecorder()

	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPropertyHandler_Delete_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "FP-001", SizeInSqm: 500}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	h := newPropertyHandler(repo, &mocks.UserRepositoryMock{})

	r := chi.NewRouter()
	r.Delete("/api/v1/properties/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/properties/"+propID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
