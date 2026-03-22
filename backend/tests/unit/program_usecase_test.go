package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	programuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/program"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestCreateProgramUseCase_Success(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		CreateFn: func(ctx context.Context, p *entity.Program) error {
			return nil
		},
	}

	uc := programuc.NewCreateProgramUseCase(mock)
	p, _ := entity.NewProgram("Test", "desc", entity.ProgramStatusActive, nil, nil)

	err := uc.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateProgramUseCase_ValidationError(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{}
	uc := programuc.NewCreateProgramUseCase(mock)

	p := &entity.Program{Name: "", Status: entity.ProgramStatusActive}
	err := uc.Execute(context.Background(), p)
	if !errors.Is(err, entity.ErrProgramNameRequired) {
		t.Errorf("expected ErrProgramNameRequired, got %v", err)
	}
}

func TestCreateProgramUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	mock := &mocks.ProgramRepositoryMock{
		CreateFn: func(ctx context.Context, p *entity.Program) error {
			return repoErr
		},
	}

	uc := programuc.NewCreateProgramUseCase(mock)
	p, _ := entity.NewProgram("Test", "desc", entity.ProgramStatusActive, nil, nil)

	err := uc.Execute(context.Background(), p)
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got %v", err)
	}
}

func TestGetProgramUseCase_Found(t *testing.T) {
	id := uuid.New()
	expected := &entity.Program{ID: id, Name: "Test", Status: entity.ProgramStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, gotID uuid.UUID) (*entity.Program, error) {
			if gotID != id {
				t.Errorf("expected ID %v, got %v", id, gotID)
			}
			return expected, nil
		},
	}

	uc := programuc.NewGetProgramUseCase(mock)
	result, err := uc.Execute(context.Background(), id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != id {
		t.Errorf("expected program with ID %v, got %v", id, result.ID)
	}
}

func TestGetProgramUseCase_NotFound(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
			return nil, nil
		},
	}

	uc := programuc.NewGetProgramUseCase(mock)
	_, err := uc.Execute(context.Background(), uuid.New())

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NOT_FOUND error, got %v", err)
	}
}

func TestListProgramsUseCase_DefaultPagination(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error) {
			if filter.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", filter.Limit)
			}
			return nil, 0, nil
		},
	}

	uc := programuc.NewListProgramsUseCase(mock)
	_, _, err := uc.Execute(context.Background(), repository.ProgramFilter{Limit: 0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListProgramsUseCase_CapsLimit(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error) {
			if filter.Limit != 100 {
				t.Errorf("expected capped limit 100, got %d", filter.Limit)
			}
			return nil, 0, nil
		},
	}

	uc := programuc.NewListProgramsUseCase(mock)
	_, _, err := uc.Execute(context.Background(), repository.ProgramFilter{Limit: 500})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteProgramUseCase_Success(t *testing.T) {
	id := uuid.New()
	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, gotID uuid.UUID) (*entity.Program, error) {
			return &entity.Program{ID: gotID, Name: "Test", Status: entity.ProgramStatusActive}, nil
		},
		DeleteFn: func(ctx context.Context, gotID uuid.UUID) error {
			if gotID != id {
				t.Errorf("expected ID %v, got %v", id, gotID)
			}
			return nil
		},
	}

	uc := programuc.NewDeleteProgramUseCase(mock)
	err := uc.Execute(context.Background(), id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteProgramUseCase_NotFound(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
			return nil, nil
		},
	}

	uc := programuc.NewDeleteProgramUseCase(mock)
	err := uc.Execute(context.Background(), uuid.New())

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NOT_FOUND error, got %v", err)
	}
}

func TestUpdateProgramUseCase_Success(t *testing.T) {
	id := uuid.New()
	existing := &entity.Program{
		ID:        id,
		Name:      "Old",
		Status:    entity.ProgramStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, gotID uuid.UUID) (*entity.Program, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Program) error {
			if p.Name != "Updated" {
				t.Errorf("expected name 'Updated', got %q", p.Name)
			}
			if p.ID != id {
				t.Errorf("expected ID %v, got %v", id, p.ID)
			}
			return nil
		},
	}

	uc := programuc.NewUpdateProgramUseCase(mock)
	updated := &entity.Program{Name: "Updated", Status: entity.ProgramStatusActive}
	err := uc.Execute(context.Background(), id, updated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateProgramUseCase_NotFound(t *testing.T) {
	mock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
			return nil, nil
		},
	}

	uc := programuc.NewUpdateProgramUseCase(mock)
	err := uc.Execute(context.Background(), uuid.New(), &entity.Program{Name: "Test", Status: entity.ProgramStatusActive})

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NOT_FOUND error, got %v", err)
	}
}
