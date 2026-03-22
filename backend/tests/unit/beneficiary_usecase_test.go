package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	beneficiaryuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/beneficiary"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func validBeneficiary() *entity.Beneficiary {
	email := "test@example.com"
	return &entity.Beneficiary{
		ID:       uuid.New(),
		FullName: "Test Person",
		Email:    &email,
		Status:   entity.BeneficiaryStatusActive,
	}
}

// --- Create ---

func TestCreateBeneficiaryUseCase_Success(t *testing.T) {
	mock := &mocks.BeneficiaryRepositoryMock{
		CreateFn: func(ctx context.Context, b *entity.Beneficiary) error { return nil },
	}
	uc := beneficiaryuc.NewCreateBeneficiaryUseCase(mock)

	b := validBeneficiary()
	if err := uc.Execute(context.Background(), b); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateBeneficiaryUseCase_ValidationError(t *testing.T) {
	mock := &mocks.BeneficiaryRepositoryMock{}
	uc := beneficiaryuc.NewCreateBeneficiaryUseCase(mock)

	b := &entity.Beneficiary{FullName: "", Status: entity.BeneficiaryStatusActive}
	err := uc.Execute(context.Background(), b)
	if !errors.Is(err, entity.ErrBeneficiaryFullNameRequired) {
		t.Errorf("expected ErrBeneficiaryFullNameRequired, got %v", err)
	}
}

func TestCreateBeneficiaryUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	mock := &mocks.BeneficiaryRepositoryMock{
		CreateFn: func(ctx context.Context, b *entity.Beneficiary) error { return repoErr },
	}
	uc := beneficiaryuc.NewCreateBeneficiaryUseCase(mock)

	b := validBeneficiary()
	err := uc.Execute(context.Background(), b)
	if err != repoErr {
		t.Errorf("expected repo error, got %v", err)
	}
}

// --- Get ---

func TestGetBeneficiaryUseCase_Success(t *testing.T) {
	b := validBeneficiary()
	mock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return b, nil },
	}
	uc := beneficiaryuc.NewGetBeneficiaryUseCase(mock)

	result, err := uc.Execute(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != b.ID {
		t.Errorf("expected ID %v, got %v", b.ID, result.ID)
	}
}

func TestGetBeneficiaryUseCase_NotFound(t *testing.T) {
	mock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return nil, nil },
	}
	uc := beneficiaryuc.NewGetBeneficiaryUseCase(mock)

	_, err := uc.Execute(context.Background(), uuid.New())
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

// --- List ---

func TestListBeneficiariesUseCase_Success(t *testing.T) {
	mock := &mocks.BeneficiaryRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
			return []*entity.Beneficiary{validBeneficiary()}, 1, nil
		},
	}
	uc := beneficiaryuc.NewListBeneficiariesUseCase(mock)

	results, total, err := uc.Execute(context.Background(), repository.BeneficiaryFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestListBeneficiariesUseCase_DefaultPagination(t *testing.T) {
	var capturedFilter repository.BeneficiaryFilter
	mock := &mocks.BeneficiaryRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := beneficiaryuc.NewListBeneficiariesUseCase(mock)

	uc.Execute(context.Background(), repository.BeneficiaryFilter{Limit: 0, Offset: -5})
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
	if capturedFilter.Offset != 0 {
		t.Errorf("expected offset 0, got %d", capturedFilter.Offset)
	}
}

func TestListBeneficiariesUseCase_MaxLimit(t *testing.T) {
	var capturedFilter repository.BeneficiaryFilter
	mock := &mocks.BeneficiaryRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := beneficiaryuc.NewListBeneficiariesUseCase(mock)

	uc.Execute(context.Background(), repository.BeneficiaryFilter{Limit: 500})
	if capturedFilter.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedFilter.Limit)
	}
}

// --- Update ---

func TestUpdateBeneficiaryUseCase_Success(t *testing.T) {
	existing := validBeneficiary()
	mock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return existing, nil },
		UpdateFn:  func(ctx context.Context, b *entity.Beneficiary) error { return nil },
	}
	uc := beneficiaryuc.NewUpdateBeneficiaryUseCase(mock)

	updated := validBeneficiary()
	updated.FullName = "Updated Name"
	if err := uc.Execute(context.Background(), existing.ID, updated); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateBeneficiaryUseCase_NotFound(t *testing.T) {
	mock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return nil, nil },
	}
	uc := beneficiaryuc.NewUpdateBeneficiaryUseCase(mock)

	err := uc.Execute(context.Background(), uuid.New(), validBeneficiary())
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

// --- Delete ---

func TestDeleteBeneficiaryUseCase_Success(t *testing.T) {
	existing := validBeneficiary()
	mock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return existing, nil },
		DeleteFn:  func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	uc := beneficiaryuc.NewDeleteBeneficiaryUseCase(mock)

	if err := uc.Execute(context.Background(), existing.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteBeneficiaryUseCase_NotFound(t *testing.T) {
	mock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return nil, nil },
	}
	uc := beneficiaryuc.NewDeleteBeneficiaryUseCase(mock)

	err := uc.Execute(context.Background(), uuid.New())
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

// --- Enroll in Program ---

func TestEnrollInProgramUseCase_Success(t *testing.T) {
	beneficiary := validBeneficiary()
	program := &entity.Program{ID: uuid.New(), Name: "Test Program", Status: entity.ProgramStatusActive}

	beneficiaryMock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn:         func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return beneficiary, nil },
		EnrollInProgramFn: func(ctx context.Context, bID, pID uuid.UUID) error { return nil },
	}
	programMock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Program, error) { return program, nil },
	}
	uc := beneficiaryuc.NewEnrollInProgramUseCase(beneficiaryMock, programMock)

	if err := uc.Execute(context.Background(), beneficiary.ID, program.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestEnrollInProgramUseCase_BeneficiaryNotFound(t *testing.T) {
	beneficiaryMock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return nil, nil },
	}
	programMock := &mocks.ProgramRepositoryMock{}
	uc := beneficiaryuc.NewEnrollInProgramUseCase(beneficiaryMock, programMock)

	err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

func TestEnrollInProgramUseCase_ProgramNotFound(t *testing.T) {
	beneficiary := validBeneficiary()
	beneficiaryMock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return beneficiary, nil },
	}
	programMock := &mocks.ProgramRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Program, error) { return nil, nil },
	}
	uc := beneficiaryuc.NewEnrollInProgramUseCase(beneficiaryMock, programMock)

	err := uc.Execute(context.Background(), beneficiary.ID, uuid.New())
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NotFound error for program, got %v", err)
	}
}

// --- Remove from Program ---

func TestRemoveFromProgramUseCase_Success(t *testing.T) {
	beneficiary := validBeneficiary()
	beneficiaryMock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn:           func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return beneficiary, nil },
		RemoveFromProgramFn: func(ctx context.Context, bID, pID uuid.UUID) error { return nil },
	}
	uc := beneficiaryuc.NewRemoveFromProgramUseCase(beneficiaryMock)

	if err := uc.Execute(context.Background(), beneficiary.ID, uuid.New()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRemoveFromProgramUseCase_BeneficiaryNotFound(t *testing.T) {
	beneficiaryMock := &mocks.BeneficiaryRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) { return nil, nil },
	}
	uc := beneficiaryuc.NewRemoveFromProgramUseCase(beneficiaryMock)

	err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected NotFound error, got %v", err)
	}
}
