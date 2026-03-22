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

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	programuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/program"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTestProgramHandler(mock *mocks.ProgramRepositoryMock) *handler.ProgramHandler {
	return handler.NewProgramHandler(
		programuc.NewCreateProgramUseCase(mock),
		programuc.NewGetProgramUseCase(mock),
		programuc.NewListProgramsUseCase(mock),
		programuc.NewUpdateProgramUseCase(mock),
		programuc.NewDeleteProgramUseCase(mock),
	)
}

func TestProgramHandler_Create_Success(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		CreateFn: func(ctx context.Context, p *entity.Program) error {
			return nil
		},
	}

	h := newTestProgramHandler(mock)
	body := `{"name":"Test Program","description":"A test","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/programs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp dto.ProgramResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Name != "Test Program" {
		t.Errorf("expected name 'Test Program', got %q", resp.Name)
	}
}

func TestProgramHandler_Create_InvalidJSON(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{}
	h := newTestProgramHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/programs", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestProgramHandler_Create_ValidationFailed(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{}
	h := newTestProgramHandler(mock)

	body := `{"name":"","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/programs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusUnprocessableEntity, rr.Code, rr.Body.String())
	}
}

func TestProgramHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	expected := &entity.Program{
		ID: id, Name: "Test", Description: "desc",
		Status: entity.ProgramStatusActive, CreatedAt: now, UpdatedAt: now,
	}

	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, gotID uuid.UUID) (*entity.Program, error) {
			return expected, nil
		},
	}

	h := newTestProgramHandler(mock)

	r := chi.NewRouter()
	r.Get("/api/v1/programs/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/programs/"+id.String(), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestProgramHandler_Get_NotFound(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
			return nil, nil
		},
	}

	h := newTestProgramHandler(mock)

	r := chi.NewRouter()
	r.Get("/api/v1/programs/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/programs/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestProgramHandler_Get_InvalidID(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{}
	h := newTestProgramHandler(mock)

	r := chi.NewRouter()
	r.Get("/api/v1/programs/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/programs/not-a-uuid", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestProgramHandler_List_Success(t *testing.T) {
	now := time.Now().UTC()
	programs := []*entity.Program{
		{ID: uuid.New(), Name: "Program 1", Status: entity.ProgramStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Name: "Program 2", Status: entity.ProgramStatusInactive, CreatedAt: now, UpdatedAt: now},
	}

	mock := &mocks.ProgramRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error) {
			return programs, 2, nil
		},
	}

	h := newTestProgramHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/programs?limit=10&offset=0", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp dto.ProgramListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 programs, got %d", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
}

func TestProgramHandler_Delete_Success(t *testing.T) {
	id := uuid.New()
	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, gotID uuid.UUID) (*entity.Program, error) {
			return &entity.Program{ID: gotID, Name: "Test", Status: entity.ProgramStatusActive}, nil
		},
		DeleteFn: func(ctx context.Context, gotID uuid.UUID) error {
			return nil
		},
	}

	h := newTestProgramHandler(mock)

	r := chi.NewRouter()
	r.Delete("/api/v1/programs/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/programs/"+id.String(), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusNoContent, rr.Code, rr.Body.String())
	}
}
