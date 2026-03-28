package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/usecase/tenant"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- CreateTenant ---

func TestCreateTenant_Success(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{
		GetByEmailFn: func(ctx context.Context, email string) (*entity.Tenant, error) {
			return nil, nil // no duplicate
		},
		CreateFn: func(ctx context.Context, ten *entity.Tenant) error {
			return nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	uc := tenant.NewCreateTenantUseCase(repo, userRepo, &mocks.AuditRepositoryMock{})
	email := "tenant@example.com"
	ten, err := entity.NewTenant("John Doe", &email, nil, nil, nil, uuid.New(), nil)
	if err != nil {
		t.Fatal(err)
	}

	err = uc.Execute(context.Background(), ten)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateTenant_LandlordNotFound(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, nil // landlord not found
		},
	}

	uc := tenant.NewCreateTenantUseCase(repo, userRepo, &mocks.AuditRepositoryMock{})
	email := "tenant@example.com"
	ten, _ := entity.NewTenant("John Doe", &email, nil, nil, nil, uuid.New(), nil)

	err := uc.Execute(context.Background(), ten)
	if err == nil {
		t.Fatal("expected error when landlord not found")
	}
}

func TestCreateTenant_DuplicateEmail(t *testing.T) {
	email := "existing@example.com"
	repo := &mocks.TenantRepositoryMock{
		GetByEmailFn: func(ctx context.Context, e string) (*entity.Tenant, error) {
			return &entity.Tenant{ID: uuid.New(), Email: &email}, nil // already exists
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	uc := tenant.NewCreateTenantUseCase(repo, userRepo, &mocks.AuditRepositoryMock{})
	ten, _ := entity.NewTenant("John Doe", &email, nil, nil, nil, uuid.New(), nil)

	err := uc.Execute(context.Background(), ten)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

// --- GetTenant ---

func TestGetTenant_Success(t *testing.T) {
	tenantID := uuid.New()
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, FullName: "John", Email: &email, LandlordID: uuid.New()}, nil
		},
	}

	uc := tenant.NewGetTenantUseCase(repo)
	result, err := uc.Execute(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != tenantID {
		t.Error("expected tenant ID to match")
	}
}

func TestGetTenant_NotFound(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}

	uc := tenant.NewGetTenantUseCase(repo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- ListTenants ---

func TestListTenants_Success(t *testing.T) {
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
			return []*entity.Tenant{{ID: uuid.New(), FullName: "John", Email: &email}}, 1, nil
		},
	}

	uc := tenant.NewListTenantsUseCase(repo)
	tenants, total, err := uc.Execute(context.Background(), repository.TenantFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(tenants) != 1 {
		t.Errorf("expected 1 tenant, got %d", len(tenants))
	}
}

func TestListTenants_DefaultLimit(t *testing.T) {
	var capturedFilter repository.TenantFilter
	repo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}

	uc := tenant.NewListTenantsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.TenantFilter{Limit: 0})
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
}

func TestListTenants_MaxLimit(t *testing.T) {
	var capturedFilter repository.TenantFilter
	repo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}

	uc := tenant.NewListTenantsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.TenantFilter{Limit: 500})
	if capturedFilter.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedFilter.Limit)
	}
}

// --- UpdateTenant ---

func TestUpdateTenant_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	email := "test@example.com"
	existingTenant := &entity.Tenant{
		ID: tenantID, FullName: "John", Email: &email, LandlordID: landlordID,
		IsActive: true, CreatedAt: time.Now().UTC(),
	}

	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return existingTenant, nil
		},
		UpdateFn: func(ctx context.Context, ten *entity.Tenant) error {
			return nil
		},
	}

	uc := tenant.NewUpdateTenantUseCase(repo, &mocks.AuditRepositoryMock{})
	updated := &entity.Tenant{FullName: "John Updated", Email: &email}
	err := uc.Execute(context.Background(), tenantID, updated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateTenant_NotFound(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}

	uc := tenant.NewUpdateTenantUseCase(repo, &mocks.AuditRepositoryMock{})
	email := "test@example.com"
	updated := &entity.Tenant{FullName: "John", Email: &email}
	err := uc.Execute(context.Background(), uuid.New(), updated)
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- DeactivateTenant ---

func TestDeactivateTenant_Success(t *testing.T) {
	tenantID := uuid.New()
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, FullName: "John", Email: &email, IsActive: true}, nil
		},
		UpdateFn: func(ctx context.Context, ten *entity.Tenant) error {
			if ten.IsActive {
				t.Error("expected IsActive to be false")
			}
			return nil
		},
	}

	uc := tenant.NewDeactivateTenantUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeactivateTenant_NotFound(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}

	uc := tenant.NewDeactivateTenantUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- DeleteTenant ---

func TestDeleteTenant_Success(t *testing.T) {
	tenantID := uuid.New()
	email := "test@example.com"
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, FullName: "John", Email: &email}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	uc := tenant.NewDeleteTenantUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteTenant_NotFound(t *testing.T) {
	repo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}

	uc := tenant.NewDeleteTenantUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}
