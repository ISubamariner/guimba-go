package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	beneficiaryuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/beneficiary"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func setupBeneficiaryHandler() (*handler.BeneficiaryHandler, *mocks.BeneficiaryRepositoryMock, *mocks.ProgramRepositoryMock) {
	beneficiaryMock := &mocks.BeneficiaryRepositoryMock{}
	programMock := &mocks.ProgramRepositoryMock{}

	createUC := beneficiaryuc.NewCreateBeneficiaryUseCase(beneficiaryMock)
	getUC := beneficiaryuc.NewGetBeneficiaryUseCase(beneficiaryMock)
	listUC := beneficiaryuc.NewListBeneficiariesUseCase(beneficiaryMock)
	updateUC := beneficiaryuc.NewUpdateBeneficiaryUseCase(beneficiaryMock)
	deleteUC := beneficiaryuc.NewDeleteBeneficiaryUseCase(beneficiaryMock)
	enrollUC := beneficiaryuc.NewEnrollInProgramUseCase(beneficiaryMock, programMock)
	removeUC := beneficiaryuc.NewRemoveFromProgramUseCase(beneficiaryMock)

	h := handler.NewBeneficiaryHandler(createUC, getUC, listUC, updateUC, deleteUC, enrollUC, removeUC)
	return h, beneficiaryMock, programMock
}

func TestBeneficiaryHandler_Create_Success(t *testing.T) {
	h, mock, _ := setupBeneficiaryHandler()
	mock.CreateFn = func(ctx context.Context, b *entity.Beneficiary) error { return nil }

	body := `{"full_name":"John Doe","email":"john@example.com","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/beneficiaries", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBeneficiaryHandler_Create_InvalidJSON(t *testing.T) {
	h, _, _ := setupBeneficiaryHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/beneficiaries", bytes.NewBufferString("{bad"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBeneficiaryHandler_Create_ValidationError(t *testing.T) {
	h, _, _ := setupBeneficiaryHandler()

	body := `{"full_name":"","email":"john@example.com","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/beneficiaries", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBeneficiaryHandler_Get_Success(t *testing.T) {
	h, mock, _ := setupBeneficiaryHandler()
	b := validBeneficiary()
	mock.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return b, nil }

	r := chi.NewRouter()
	r.Get("/api/v1/beneficiaries/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/beneficiaries/"+b.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBeneficiaryHandler_Get_NotFound(t *testing.T) {
	h, mock, _ := setupBeneficiaryHandler()
	mock.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return nil, nil }

	r := chi.NewRouter()
	r.Get("/api/v1/beneficiaries/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/beneficiaries/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestBeneficiaryHandler_List_Success(t *testing.T) {
	h, mock, _ := setupBeneficiaryHandler()
	mock.ListFn = func(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
		return []*entity.Beneficiary{validBeneficiary()}, 1, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/beneficiaries?status=active&limit=10", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

func TestBeneficiaryHandler_Delete_Success(t *testing.T) {
	h, mock, _ := setupBeneficiaryHandler()
	b := validBeneficiary()
	mock.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return b, nil }
	mock.DeleteFn = func(ctx context.Context, id uuid.UUID) error { return nil }

	r := chi.NewRouter()
	r.Delete("/api/v1/beneficiaries/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/beneficiaries/"+b.ID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBeneficiaryHandler_EnrollInProgram_Success(t *testing.T) {
	h, beneficiaryMock, programMock := setupBeneficiaryHandler()
	b := validBeneficiary()
	p := &entity.Program{ID: uuid.New(), Name: "Test Program", Status: entity.ProgramStatusActive}

	beneficiaryMock.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return b, nil }
	beneficiaryMock.EnrollInProgramFn = func(ctx context.Context, bID, pID uuid.UUID) error { return nil }
	programMock.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entity.Program, error) { return p, nil }

	r := chi.NewRouter()
	r.Post("/api/v1/beneficiaries/{id}/programs", h.EnrollInProgram)

	body := `{"program_id":"` + p.ID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/beneficiaries/"+b.ID.String()+"/programs", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
