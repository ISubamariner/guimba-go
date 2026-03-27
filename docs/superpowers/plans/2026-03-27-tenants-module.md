# Tenants Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the Tenants domain module — entity, repository, use cases, persistence, DTOs, handler, routes, and tests — following the existing Clean Architecture patterns.

**Architecture:** Faithful port of Tenant entity from the original Python system. Tenants belong to a landlord (User). Multi-tenant data isolation enforced at the use case layer via ownership checks. Address is a reusable value object stored as flat columns.

**Tech Stack:** Go 1.26+, Chi v5, pgx v5, go-playground/validator v10, google/uuid

**Spec:** `docs/superpowers/specs/2026-03-27-tenants-module-design.md`

---

## File Map

| Action | File | Responsibility |
|:---|:---|:---|
| Create | `backend/internal/domain/entity/address.go` | Address value object |
| Create | `backend/internal/domain/entity/tenant.go` | Tenant entity + validation |
| Modify | `backend/internal/domain/entity/errors.go` | Add tenant domain errors |
| Create | `backend/internal/domain/repository/tenant_repository.go` | TenantRepository interface + TenantFilter |
| Create | `backend/internal/usecase/tenant/create_tenant.go` | CreateTenant use case |
| Create | `backend/internal/usecase/tenant/get_tenant.go` | GetTenant use case |
| Create | `backend/internal/usecase/tenant/list_tenants.go` | ListTenants use case |
| Create | `backend/internal/usecase/tenant/update_tenant.go` | UpdateTenant use case |
| Create | `backend/internal/usecase/tenant/deactivate_tenant.go` | DeactivateTenant use case |
| Create | `backend/internal/usecase/tenant/delete_tenant.go` | DeleteTenant use case |
| Create | `backend/internal/delivery/http/dto/tenant_dto.go` | Request/response DTOs + converters |
| Create | `backend/internal/delivery/http/handler/tenant_handler.go` | HTTP handlers with Swagger |
| Create | `backend/internal/infrastructure/persistence/pg/tenant_repo_pg.go` | PostgreSQL repository |
| Create | `backend/migrations/000007_create_tenants.up.sql` | Tenants table |
| Create | `backend/migrations/000007_create_tenants.down.sql` | Drop tenants table |
| Modify | `backend/internal/delivery/http/router/router.go` | Add Tenant field + routes |
| Modify | `backend/cmd/server/main.go` | Wire tenant module |
| Create | `backend/tests/mocks/tenant_repository_mock.go` | Manual mock |
| Create | `backend/tests/unit/tenant_entity_test.go` | Entity validation tests |
| Create | `backend/tests/unit/tenant_usecase_test.go` | Use case tests |
| Create | `backend/tests/unit/tenant_handler_test.go` | Handler tests |

---

## Task 1: Domain — Address Value Object

**Files:**
- Create: `backend/internal/domain/entity/address.go`

- [ ] **Step 1: Create Address value object**

```go
package entity

// Address represents a physical address (value object, reusable across entities).
type Address struct {
	Street        string `json:"street"`
	City          string `json:"city"`
	StateOrRegion string `json:"state_or_region"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country"`
}

// NewAddress creates an Address with defaults.
func NewAddress(street, city, stateOrRegion, postalCode, country string) *Address {
	if country == "" {
		country = "Philippines"
	}
	return &Address{
		Street:        street,
		City:          city,
		StateOrRegion: stateOrRegion,
		PostalCode:    postalCode,
		Country:       country,
	}
}

// FullAddress returns the formatted address string.
func (a *Address) FullAddress() string {
	s := a.Street + ", " + a.City + ", " + a.StateOrRegion
	if a.PostalCode != "" {
		s += ", " + a.PostalCode
	}
	s += ", " + a.Country
	return s
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/domain/entity/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/entity/address.go
git commit -m "feat(tenant): add Address value object"
```

---

## Task 2: Domain — Tenant Entity + Errors

**Files:**
- Create: `backend/internal/domain/entity/tenant.go`
- Modify: `backend/internal/domain/entity/errors.go`
- Create: `backend/tests/unit/tenant_entity_test.go`

- [ ] **Step 1: Add tenant domain errors to errors.go**

Append this block to `backend/internal/domain/entity/errors.go`:

```go
// Domain errors for Tenant entity.
var (
	ErrTenantFullNameRequired = errors.New("tenant full name is required")
	ErrTenantFullNameTooLong  = errors.New("tenant full name must be 255 characters or less")
	ErrTenantContactRequired  = errors.New("tenant must have at least one contact method (email or phone)")
	ErrTenantEmailExists      = errors.New("a tenant with this email already exists")
)
```

- [ ] **Step 2: Write the failing entity tests**

Create `backend/tests/unit/tenant_entity_test.go`:

```go
package unit

import (
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/google/uuid"
)

func TestNewTenant_ValidWithEmail(t *testing.T) {
	email := "tenant@example.com"
	landlordID := uuid.New()
	tenant, err := entity.NewTenant("John Doe", &email, nil, nil, nil, landlordID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.FullName != "John Doe" {
		t.Errorf("expected FullName 'John Doe', got %q", tenant.FullName)
	}
	if tenant.ID == (uuid.UUID{}) {
		t.Error("expected non-zero UUID")
	}
	if tenant.LandlordID != landlordID {
		t.Error("expected landlord ID to match")
	}
	if !tenant.IsActive {
		t.Error("expected IsActive to default to true")
	}
}

func TestNewTenant_ValidWithPhone(t *testing.T) {
	phone := "+639171234567"
	tenant, err := entity.NewTenant("Jane Doe", nil, &phone, nil, nil, uuid.New(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *tenant.PhoneNumber != phone {
		t.Errorf("expected phone %q, got %q", phone, *tenant.PhoneNumber)
	}
}

func TestNewTenant_NameRequired(t *testing.T) {
	email := "test@example.com"
	_, err := entity.NewTenant("", &email, nil, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantFullNameRequired {
		t.Errorf("expected ErrTenantFullNameRequired, got %v", err)
	}
}

func TestNewTenant_NameTooLong(t *testing.T) {
	email := "test@example.com"
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := entity.NewTenant(string(longName), &email, nil, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantFullNameTooLong {
		t.Errorf("expected ErrTenantFullNameTooLong, got %v", err)
	}
}

func TestNewTenant_ContactRequired(t *testing.T) {
	_, err := entity.NewTenant("John", nil, nil, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantContactRequired {
		t.Errorf("expected ErrTenantContactRequired, got %v", err)
	}
}

func TestNewTenant_EmptyContactStrings(t *testing.T) {
	empty := ""
	_, err := entity.NewTenant("John", &empty, &empty, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantContactRequired {
		t.Errorf("expected ErrTenantContactRequired with empty strings, got %v", err)
	}
}

func TestNewTenant_WithAddress(t *testing.T) {
	email := "test@example.com"
	addr := entity.NewAddress("123 Main St", "Guimba", "Nueva Ecija", "3115", "")
	tenant, err := entity.NewTenant("John", &email, nil, nil, addr, uuid.New(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.Address == nil {
		t.Fatal("expected address to be set")
	}
	if tenant.Address.Country != "Philippines" {
		t.Errorf("expected default country 'Philippines', got %q", tenant.Address.Country)
	}
}

func TestNewTenant_WithAllFields(t *testing.T) {
	email := "john@example.com"
	phone := "+639171234567"
	nid := "NID-123"
	notes := "Good tenant"
	addr := entity.NewAddress("123 Main St", "Guimba", "Nueva Ecija", "3115", "Philippines")

	tenant, err := entity.NewTenant("John Doe", &email, &phone, &nid, addr, uuid.New(), &notes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *tenant.Email != email {
		t.Error("email mismatch")
	}
	if *tenant.PhoneNumber != phone {
		t.Error("phone mismatch")
	}
	if *tenant.NationalID != nid {
		t.Error("national_id mismatch")
	}
	if *tenant.Notes != notes {
		t.Error("notes mismatch")
	}
	if tenant.Address.Street != "123 Main St" {
		t.Error("address street mismatch")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd backend && go test ./tests/unit/ -run TestNewTenant -v`
Expected: FAIL — `entity.NewTenant` not defined

- [ ] **Step 4: Create Tenant entity**

Create `backend/internal/domain/entity/tenant.go`:

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a person who owes money to a landlord.
type Tenant struct {
	ID          uuid.UUID  `json:"id"`
	FullName    string     `json:"full_name"`
	Email       *string    `json:"email,omitempty"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	NationalID  *string    `json:"national_id,omitempty"`
	Address     *Address   `json:"address,omitempty"`
	LandlordID  uuid.UUID  `json:"landlord_id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewTenant creates a new Tenant with generated ID and defaults.
func NewTenant(fullName string, email, phoneNumber, nationalID *string, address *Address, landlordID uuid.UUID, notes *string) (*Tenant, error) {
	t := &Tenant{
		ID:          uuid.New(),
		FullName:    fullName,
		Email:       email,
		PhoneNumber: phoneNumber,
		NationalID:  nationalID,
		Address:     address,
		LandlordID:  landlordID,
		IsActive:    true,
		Notes:       notes,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t, nil
}

// Validate checks business rules for a Tenant.
func (t *Tenant) Validate() error {
	if t.FullName == "" {
		return ErrTenantFullNameRequired
	}
	if len(t.FullName) > 255 {
		return ErrTenantFullNameTooLong
	}
	hasEmail := t.Email != nil && *t.Email != ""
	hasPhone := t.PhoneNumber != nil && *t.PhoneNumber != ""
	if !hasEmail && !hasPhone {
		return ErrTenantContactRequired
	}
	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./tests/unit/ -run TestNewTenant -v`
Expected: All 8 tests PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/entity/tenant.go backend/internal/domain/entity/errors.go backend/tests/unit/tenant_entity_test.go
git commit -m "feat(tenant): add Tenant entity, Address value object, and domain errors with tests"
```

---

## Task 3: Domain — Repository Interface

**Files:**
- Create: `backend/internal/domain/repository/tenant_repository.go`

- [ ] **Step 1: Create the repository interface**

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// TenantFilter holds optional filters for listing tenants.
type TenantFilter struct {
	LandlordID *uuid.UUID
	IsActive   *bool
	Search     *string
	Limit      int
	Offset     int
}

// TenantRepository defines the interface for tenant persistence operations.
type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetByEmail(ctx context.Context, email string) (*entity.Tenant, error)
	List(ctx context.Context, filter TenantFilter) ([]*entity.Tenant, int, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/domain/repository/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/repository/tenant_repository.go
git commit -m "feat(tenant): add TenantRepository interface"
```

---

## Task 4: Mock — Tenant Repository Mock

**Files:**
- Create: `backend/tests/mocks/tenant_repository_mock.go`

- [ ] **Step 1: Create the mock**

```go
package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// TenantRepositoryMock is a test mock for repository.TenantRepository.
type TenantRepositoryMock struct {
	CreateFn     func(ctx context.Context, tenant *entity.Tenant) error
	GetByIDFn    func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetByEmailFn func(ctx context.Context, email string) (*entity.Tenant, error)
	ListFn       func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error)
	UpdateFn     func(ctx context.Context, tenant *entity.Tenant) error
	DeleteFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *TenantRepositoryMock) Create(ctx context.Context, tenant *entity.Tenant) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tenant)
	}
	return nil
}

func (m *TenantRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *TenantRepositoryMock) GetByEmail(ctx context.Context, email string) (*entity.Tenant, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *TenantRepositoryMock) List(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *TenantRepositoryMock) Update(ctx context.Context, tenant *entity.Tenant) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, tenant)
	}
	return nil
}

func (m *TenantRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./tests/mocks/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/tests/mocks/tenant_repository_mock.go
git commit -m "feat(tenant): add TenantRepository mock"
```

---

## Task 5: Use Cases — Create, Get, List

**Files:**
- Create: `backend/internal/usecase/tenant/create_tenant.go`
- Create: `backend/internal/usecase/tenant/get_tenant.go`
- Create: `backend/internal/usecase/tenant/list_tenants.go`
- Create: `backend/tests/unit/tenant_usecase_test.go` (partial — tests for these 3)

- [ ] **Step 1: Write failing use case tests (create, get, list)**

Create `backend/tests/unit/tenant_usecase_test.go`:

```go
package unit

import (
	"context"
	"testing"

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

	uc := tenant.NewCreateTenantUseCase(repo, userRepo)
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

	uc := tenant.NewCreateTenantUseCase(repo, userRepo)
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

	uc := tenant.NewCreateTenantUseCase(repo, userRepo)
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./tests/unit/ -run "TestCreateTenant|TestGetTenant|TestListTenants" -v`
Expected: FAIL — packages don't exist yet

- [ ] **Step 3: Implement CreateTenant use case**

Create `backend/internal/usecase/tenant/create_tenant.go`:

```go
package tenant

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// CreateTenantUseCase handles creating a new tenant.
type CreateTenantUseCase struct {
	repo     repository.TenantRepository
	userRepo repository.UserRepository
}

// NewCreateTenantUseCase creates a new CreateTenantUseCase.
func NewCreateTenantUseCase(repo repository.TenantRepository, userRepo repository.UserRepository) *CreateTenantUseCase {
	return &CreateTenantUseCase{repo: repo, userRepo: userRepo}
}

// Execute creates a new tenant after validating the landlord exists and email is unique.
func (uc *CreateTenantUseCase) Execute(ctx context.Context, tenant *entity.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}

	// Validate landlord exists
	landlord, err := uc.userRepo.GetByID(ctx, tenant.LandlordID)
	if err != nil {
		return err
	}
	if landlord == nil {
		return apperror.NewNotFound("User", tenant.LandlordID)
	}

	// Check email uniqueness among tenants
	if tenant.Email != nil && *tenant.Email != "" {
		existing, err := uc.repo.GetByEmail(ctx, *tenant.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			return apperror.NewConflict(entity.ErrTenantEmailExists.Error())
		}
	}

	return uc.repo.Create(ctx, tenant)
}
```

- [ ] **Step 4: Implement GetTenant use case**

Create `backend/internal/usecase/tenant/get_tenant.go`:

```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// GetTenantUseCase handles retrieving a single tenant by ID.
type GetTenantUseCase struct {
	repo repository.TenantRepository
}

// NewGetTenantUseCase creates a new GetTenantUseCase.
func NewGetTenantUseCase(repo repository.TenantRepository) *GetTenantUseCase {
	return &GetTenantUseCase{repo: repo}
}

// Execute retrieves a tenant by ID.
func (uc *GetTenantUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, apperror.NewNotFound("Tenant", id)
	}
	return tenant, nil
}
```

- [ ] **Step 5: Implement ListTenants use case**

Create `backend/internal/usecase/tenant/list_tenants.go`:

```go
package tenant

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ListTenantsUseCase handles listing tenants with filtering and pagination.
type ListTenantsUseCase struct {
	repo repository.TenantRepository
}

// NewListTenantsUseCase creates a new ListTenantsUseCase.
func NewListTenantsUseCase(repo repository.TenantRepository) *ListTenantsUseCase {
	return &ListTenantsUseCase{repo: repo}
}

// Execute returns a filtered, paginated list of tenants and the total count.
func (uc *ListTenantsUseCase) Execute(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return uc.repo.List(ctx, filter)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend && go test ./tests/unit/ -run "TestCreateTenant|TestGetTenant|TestListTenants" -v`
Expected: All 8 tests PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/usecase/tenant/ backend/tests/unit/tenant_usecase_test.go
git commit -m "feat(tenant): add Create, Get, List use cases with tests"
```

---

## Task 6: Use Cases — Update, Deactivate, Delete

**Files:**
- Create: `backend/internal/usecase/tenant/update_tenant.go`
- Create: `backend/internal/usecase/tenant/deactivate_tenant.go`
- Create: `backend/internal/usecase/tenant/delete_tenant.go`
- Modify: `backend/tests/unit/tenant_usecase_test.go` (append tests)

- [ ] **Step 1: Append failing tests to tenant_usecase_test.go**

Append to `backend/tests/unit/tenant_usecase_test.go`:

```go
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

	uc := tenant.NewUpdateTenantUseCase(repo)
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

	uc := tenant.NewUpdateTenantUseCase(repo)
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

	uc := tenant.NewDeactivateTenantUseCase(repo)
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

	uc := tenant.NewDeactivateTenantUseCase(repo)
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

	uc := tenant.NewDeleteTenantUseCase(repo)
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

	uc := tenant.NewDeleteTenantUseCase(repo)
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}
```

Note: Add `"time"` to the imports at the top of the file.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./tests/unit/ -run "TestUpdateTenant|TestDeactivateTenant|TestDeleteTenant" -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Implement UpdateTenant use case**

Create `backend/internal/usecase/tenant/update_tenant.go`:

```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// UpdateTenantUseCase handles updating an existing tenant.
type UpdateTenantUseCase struct {
	repo repository.TenantRepository
}

// NewUpdateTenantUseCase creates a new UpdateTenantUseCase.
func NewUpdateTenantUseCase(repo repository.TenantRepository) *UpdateTenantUseCase {
	return &UpdateTenantUseCase{repo: repo}
}

// Execute updates a tenant after verifying it exists.
func (uc *UpdateTenantUseCase) Execute(ctx context.Context, id uuid.UUID, tenant *entity.Tenant) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	tenant.ID = id
	tenant.CreatedAt = existing.CreatedAt
	tenant.LandlordID = existing.LandlordID

	if err := tenant.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, tenant)
}
```

- [ ] **Step 4: Implement DeactivateTenant use case**

Create `backend/internal/usecase/tenant/deactivate_tenant.go`:

```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeactivateTenantUseCase handles deactivating a tenant.
type DeactivateTenantUseCase struct {
	repo repository.TenantRepository
}

// NewDeactivateTenantUseCase creates a new DeactivateTenantUseCase.
func NewDeactivateTenantUseCase(repo repository.TenantRepository) *DeactivateTenantUseCase {
	return &DeactivateTenantUseCase{repo: repo}
}

// Execute deactivates a tenant by setting IsActive to false.
func (uc *DeactivateTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	existing.IsActive = false
	return uc.repo.Update(ctx, existing)
}
```

- [ ] **Step 5: Implement DeleteTenant use case**

Create `backend/internal/usecase/tenant/delete_tenant.go`:

```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeleteTenantUseCase handles soft-deleting a tenant.
type DeleteTenantUseCase struct {
	repo repository.TenantRepository
}

// NewDeleteTenantUseCase creates a new DeleteTenantUseCase.
func NewDeleteTenantUseCase(repo repository.TenantRepository) *DeleteTenantUseCase {
	return &DeleteTenantUseCase{repo: repo}
}

// Execute soft-deletes a tenant by ID.
func (uc *DeleteTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	return uc.repo.Delete(ctx, id)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend && go test ./tests/unit/ -run "TestUpdateTenant|TestDeactivateTenant|TestDeleteTenant" -v`
Expected: All 6 tests PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/usecase/tenant/ backend/tests/unit/tenant_usecase_test.go
git commit -m "feat(tenant): add Update, Deactivate, Delete use cases with tests"
```

---

## Task 7: DTOs

**Files:**
- Create: `backend/internal/delivery/http/dto/tenant_dto.go`

- [ ] **Step 1: Create tenant DTOs**

```go
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// AddressDTO is the request/response shape for an address.
type AddressDTO struct {
	Street        string `json:"street" validate:"required"`
	City          string `json:"city" validate:"required"`
	StateOrRegion string `json:"state_or_region" validate:"required"`
	PostalCode    string `json:"postal_code,omitempty" validate:"omitempty"`
	Country       string `json:"country,omitempty" validate:"omitempty"`
}

// CreateTenantRequest is the request body for creating a tenant.
type CreateTenantRequest struct {
	FullName    string      `json:"full_name" validate:"required,max=255"`
	Email       *string     `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber *string     `json:"phone_number" validate:"omitempty,max=50"`
	NationalID  *string     `json:"national_id" validate:"omitempty,max=100"`
	Address     *AddressDTO `json:"address" validate:"omitempty"`
	Notes       *string     `json:"notes" validate:"omitempty"`
}

// UpdateTenantRequest is the request body for updating a tenant.
type UpdateTenantRequest struct {
	FullName    string      `json:"full_name" validate:"required,max=255"`
	Email       *string     `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber *string     `json:"phone_number" validate:"omitempty,max=50"`
	NationalID  *string     `json:"national_id" validate:"omitempty,max=100"`
	Address     *AddressDTO `json:"address" validate:"omitempty"`
	Notes       *string     `json:"notes" validate:"omitempty"`
}

// TenantResponse is the response body for a single tenant.
type TenantResponse struct {
	ID          uuid.UUID   `json:"id"`
	FullName    string      `json:"full_name"`
	Email       *string     `json:"email,omitempty"`
	PhoneNumber *string     `json:"phone_number,omitempty"`
	NationalID  *string     `json:"national_id,omitempty"`
	Address     *AddressDTO `json:"address,omitempty"`
	LandlordID  uuid.UUID   `json:"landlord_id"`
	IsActive    bool        `json:"is_active"`
	Notes       *string     `json:"notes,omitempty"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// TenantListResponse is the response body for a list of tenants.
type TenantListResponse struct {
	Data   []TenantResponse `json:"data"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// ToEntity converts a CreateTenantRequest to a domain entity.
func (r *CreateTenantRequest) ToEntity(landlordID uuid.UUID) (*entity.Tenant, error) {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	return entity.NewTenant(r.FullName, r.Email, r.PhoneNumber, r.NationalID, addr, landlordID, r.Notes)
}

// ToEntity converts an UpdateTenantRequest to a partial domain entity.
func (r *UpdateTenantRequest) ToEntity() *entity.Tenant {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	return &entity.Tenant{
		FullName:    r.FullName,
		Email:       r.Email,
		PhoneNumber: r.PhoneNumber,
		NationalID:  r.NationalID,
		Address:     addr,
		Notes:       r.Notes,
	}
}

// NewTenantResponse creates a TenantResponse from a domain entity.
func NewTenantResponse(t *entity.Tenant) TenantResponse {
	resp := TenantResponse{
		ID:          t.ID,
		FullName:    t.FullName,
		Email:       t.Email,
		PhoneNumber: t.PhoneNumber,
		NationalID:  t.NationalID,
		LandlordID:  t.LandlordID,
		IsActive:    t.IsActive,
		Notes:       t.Notes,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
	if t.Address != nil {
		resp.Address = &AddressDTO{
			Street:        t.Address.Street,
			City:          t.Address.City,
			StateOrRegion: t.Address.StateOrRegion,
			PostalCode:    t.Address.PostalCode,
			Country:       t.Address.Country,
		}
	}
	return resp
}

// NewTenantListResponse creates a TenantListResponse from domain entities.
func NewTenantListResponse(tenants []*entity.Tenant, total, limit, offset int) TenantListResponse {
	data := make([]TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		data = append(data, NewTenantResponse(t))
	}
	return TenantListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/delivery/http/dto/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/delivery/http/dto/tenant_dto.go
git commit -m "feat(tenant): add tenant DTOs with Address support"
```

---

## Task 8: HTTP Handler + Tests

**Files:**
- Create: `backend/internal/delivery/http/handler/tenant_handler.go`
- Create: `backend/tests/unit/tenant_handler_test.go`

- [ ] **Step 1: Write failing handler tests**

Create `backend/tests/unit/tenant_handler_test.go`:

```go
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
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./tests/unit/ -run "TestTenantHandler" -v`
Expected: FAIL — `handler.TenantHandler` not defined

- [ ] **Step 3: Create tenant handler**

Create `backend/internal/delivery/http/handler/tenant_handler.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	tenantuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/tenant"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

// TenantHandler handles HTTP requests for tenants.
type TenantHandler struct {
	createUC     *tenantuc.CreateTenantUseCase
	getUC        *tenantuc.GetTenantUseCase
	listUC       *tenantuc.ListTenantsUseCase
	updateUC     *tenantuc.UpdateTenantUseCase
	deactivateUC *tenantuc.DeactivateTenantUseCase
	deleteUC     *tenantuc.DeleteTenantUseCase
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(
	createUC *tenantuc.CreateTenantUseCase,
	getUC *tenantuc.GetTenantUseCase,
	listUC *tenantuc.ListTenantsUseCase,
	updateUC *tenantuc.UpdateTenantUseCase,
	deactivateUC *tenantuc.DeactivateTenantUseCase,
	deleteUC *tenantuc.DeleteTenantUseCase,
) *TenantHandler {
	return &TenantHandler{
		createUC:     createUC,
		getUC:        getUC,
		listUC:       listUC,
		updateUC:     updateUC,
		deactivateUC: deactivateUC,
		deleteUC:     deleteUC,
	}
}

// Create godoc
// @Summary      Create a tenant
// @Description  Creates a new tenant belonging to the authenticated landlord
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateTenantRequest  true  "Tenant data"
// @Success      201   {object}  dto.TenantResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      409   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants [post]
func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	tenant, err := req.ToEntity(landlordID)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	if err := h.createUC.Execute(r.Context(), tenant); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewTenantResponse(tenant))
}

// Get godoc
// @Summary      Get a tenant
// @Description  Retrieves a tenant by ID
// @Tags         tenants
// @Produce      json
// @Param        id   path      string  true  "Tenant ID (UUID)"
// @Success      200  {object}  dto.TenantResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id} [get]
func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	tenant, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantResponse(tenant))
}

// List godoc
// @Summary      List tenants
// @Description  Returns a paginated list of tenants with optional filtering
// @Tags         tenants
// @Produce      json
// @Param        landlord_id  query     string  false  "Filter by landlord ID"
// @Param        is_active    query     bool    false  "Filter by active status"
// @Param        search       query     string  false  "Search by name, email, phone"
// @Param        limit        query     int     false  "Page size (default 20, max 100)"
// @Param        offset       query     int     false  "Offset (default 0)"
// @Success      200          {object}  dto.TenantListResponse
// @Failure      500          {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants [get]
func (h *TenantHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.TenantFilter{
		Limit:  20,
		Offset: 0,
	}

	if s := r.URL.Query().Get("landlord_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid landlord_id filter"))
			return
		}
		filter.LandlordID = &id
	}

	if s := r.URL.Query().Get("is_active"); s != "" {
		v := s == "true"
		filter.IsActive = &v
	}

	if s := r.URL.Query().Get("search"); s != "" {
		filter.Search = &s
	}

	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Limit = v
		}
	}

	if s := r.URL.Query().Get("offset"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Offset = v
		}
	}

	tenants, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantListResponse(tenants, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a tenant
// @Description  Updates an existing tenant by ID
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Tenant ID (UUID)"
// @Param        body  body      dto.UpdateTenantRequest  true  "Updated tenant data"
// @Success      200   {object}  dto.TenantResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id} [put]
func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	tenant := req.ToEntity()

	if err := h.updateUC.Execute(r.Context(), id, tenant); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantResponse(updated))
}

// Deactivate godoc
// @Summary      Deactivate a tenant
// @Description  Sets a tenant's is_active to false
// @Tags         tenants
// @Produce      json
// @Param        id   path      string  true  "Tenant ID (UUID)"
// @Success      200  {object}  dto.TenantResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id}/deactivate [put]
func (h *TenantHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	if err := h.deactivateUC.Execute(r.Context(), id); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantResponse(updated))
}

// Delete godoc
// @Summary      Delete a tenant
// @Description  Soft-deletes a tenant by ID
// @Tags         tenants
// @Produce      json
// @Param        id   path      string  true  "Tenant ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id} [delete]
func (h *TenantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleTenantDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	if errors.Is(err, entity.ErrTenantFullNameRequired) ||
		errors.Is(err, entity.ErrTenantFullNameTooLong) ||
		errors.Is(err, entity.ErrTenantContactRequired) {
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
		return
	}

	if errors.Is(err, entity.ErrTenantEmailExists) {
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
		return
	}

	slog.Error("unhandled error in tenant handler", "error", err)
	apperror.WriteError(w, apperror.NewInternal(err))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./tests/unit/ -run "TestTenantHandler" -v`
Expected: All 8 tests PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/handler/tenant_handler.go backend/tests/unit/tenant_handler_test.go
git commit -m "feat(tenant): add HTTP handler with Swagger annotations and tests"
```

---

## Task 9: Database Migration

**Files:**
- Create: `backend/migrations/000007_create_tenants.up.sql`
- Create: `backend/migrations/000007_create_tenants.down.sql`

- [ ] **Step 1: Create up migration**

```sql
-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone_number VARCHAR(50),
    national_id VARCHAR(100),
    address_street VARCHAR(255),
    address_city VARCHAR(255),
    address_state_or_region VARCHAR(255),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100) DEFAULT 'Philippines',
    landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Trigger for auto-updating updated_at
CREATE TRIGGER set_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Indexes
CREATE INDEX idx_tenants_landlord_id ON tenants (landlord_id);
CREATE UNIQUE INDEX idx_tenants_email ON tenants (email) WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tenants_is_active ON tenants (is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_tenants_full_name ON tenants USING gin (full_name gin_trgm_ops);
```

- [ ] **Step 2: Create down migration**

```sql
DROP TRIGGER IF EXISTS set_tenants_updated_at ON tenants;
DROP TABLE IF EXISTS tenants;
```

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/000007_create_tenants.up.sql backend/migrations/000007_create_tenants.down.sql
git commit -m "feat(tenant): add database migration for tenants table"
```

---

## Task 10: PostgreSQL Repository

**Files:**
- Create: `backend/internal/infrastructure/persistence/pg/tenant_repo_pg.go`

- [ ] **Step 1: Create PG repository implementation**

```go
package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// TenantRepoPG implements repository.TenantRepository using PostgreSQL.
type TenantRepoPG struct {
	pool *pgxpool.Pool
}

// NewTenantRepoPG creates a new PostgreSQL tenant repository.
func NewTenantRepoPG(pool *pgxpool.Pool) *TenantRepoPG {
	return &TenantRepoPG{pool: pool}
}

func (r *TenantRepoPG) Create(ctx context.Context, t *entity.Tenant) error {
	query := `
		INSERT INTO tenants (id, full_name, email, phone_number, national_id,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			landlord_id, user_id, is_active, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	var street, city, stateOrRegion, postalCode, country *string
	if t.Address != nil {
		street = &t.Address.Street
		city = &t.Address.City
		stateOrRegion = &t.Address.StateOrRegion
		if t.Address.PostalCode != "" {
			postalCode = &t.Address.PostalCode
		}
		country = &t.Address.Country
	}

	_, err := r.pool.Exec(ctx, query,
		t.ID, t.FullName, t.Email, t.PhoneNumber, t.NationalID,
		street, city, stateOrRegion, postalCode, country,
		t.LandlordID, t.UserID, t.IsActive, t.Notes, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *TenantRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	query := `
		SELECT id, full_name, email, phone_number, national_id,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			landlord_id, user_id, is_active, notes, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL`

	return r.scanTenant(r.pool.QueryRow(ctx, query, id))
}

func (r *TenantRepoPG) GetByEmail(ctx context.Context, email string) (*entity.Tenant, error) {
	query := `
		SELECT id, full_name, email, phone_number, national_id,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			landlord_id, user_id, is_active, notes, created_at, updated_at, deleted_at
		FROM tenants
		WHERE email = $1 AND deleted_at IS NULL`

	return r.scanTenant(r.pool.QueryRow(ctx, query, email))
}

func (r *TenantRepoPG) List(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "t.deleted_at IS NULL")

	if filter.LandlordID != nil {
		conditions = append(conditions, fmt.Sprintf("t.landlord_id = $%d", argIdx))
		args = append(args, *filter.LandlordID)
		argIdx++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("t.is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(t.full_name ILIKE $%d OR t.email ILIKE $%d OR t.phone_number ILIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM tenants t " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.full_name, t.email, t.phone_number, t.national_id,
			t.address_street, t.address_city, t.address_state_or_region, t.address_postal_code, t.address_country,
			t.landlord_id, t.user_id, t.is_active, t.notes, t.created_at, t.updated_at, t.deleted_at
		FROM tenants t %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []*entity.Tenant
	for rows.Next() {
		t, err := r.scanTenantRow(rows)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, t)
	}

	return tenants, total, rows.Err()
}

func (r *TenantRepoPG) Update(ctx context.Context, t *entity.Tenant) error {
	query := `
		UPDATE tenants
		SET full_name = $1, email = $2, phone_number = $3, national_id = $4,
			address_street = $5, address_city = $6, address_state_or_region = $7,
			address_postal_code = $8, address_country = $9,
			is_active = $10, notes = $11, updated_at = $12
		WHERE id = $13 AND deleted_at IS NULL`

	var street, city, stateOrRegion, postalCode, country *string
	if t.Address != nil {
		street = &t.Address.Street
		city = &t.Address.City
		stateOrRegion = &t.Address.StateOrRegion
		if t.Address.PostalCode != "" {
			postalCode = &t.Address.PostalCode
		}
		country = &t.Address.Country
	}

	_, err := r.pool.Exec(ctx, query,
		t.FullName, t.Email, t.PhoneNumber, t.NationalID,
		street, city, stateOrRegion, postalCode, country,
		t.IsActive, t.Notes, time.Now().UTC(), t.ID,
	)
	return err
}

func (r *TenantRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenants SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

// scanTenant scans a single row into a Tenant entity.
func (r *TenantRepoPG) scanTenant(row pgx.Row) (*entity.Tenant, error) {
	t := &entity.Tenant{}
	var street, city, stateOrRegion, postalCode, country *string

	err := row.Scan(
		&t.ID, &t.FullName, &t.Email, &t.PhoneNumber, &t.NationalID,
		&street, &city, &stateOrRegion, &postalCode, &country,
		&t.LandlordID, &t.UserID, &t.IsActive, &t.Notes, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if street != nil || city != nil {
		pc := ""
		if postalCode != nil {
			pc = *postalCode
		}
		c := "Philippines"
		if country != nil {
			c = *country
		}
		s := ""
		if street != nil {
			s = *street
		}
		ci := ""
		if city != nil {
			ci = *city
		}
		sr := ""
		if stateOrRegion != nil {
			sr = *stateOrRegion
		}
		t.Address = &entity.Address{
			Street: s, City: ci, StateOrRegion: sr, PostalCode: pc, Country: c,
		}
	}

	return t, nil
}

// scanTenantRow scans a rows.Next() result into a Tenant entity.
func (r *TenantRepoPG) scanTenantRow(rows pgx.Rows) (*entity.Tenant, error) {
	t := &entity.Tenant{}
	var street, city, stateOrRegion, postalCode, country *string

	err := rows.Scan(
		&t.ID, &t.FullName, &t.Email, &t.PhoneNumber, &t.NationalID,
		&street, &city, &stateOrRegion, &postalCode, &country,
		&t.LandlordID, &t.UserID, &t.IsActive, &t.Notes, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	if street != nil || city != nil {
		pc := ""
		if postalCode != nil {
			pc = *postalCode
		}
		c := "Philippines"
		if country != nil {
			c = *country
		}
		s := ""
		if street != nil {
			s = *street
		}
		ci := ""
		if city != nil {
			ci = *city
		}
		sr := ""
		if stateOrRegion != nil {
			sr = *stateOrRegion
		}
		t.Address = &entity.Address{
			Street: s, City: ci, StateOrRegion: sr, PostalCode: pc, Country: c,
		}
	}

	return t, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/infrastructure/persistence/pg/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/infrastructure/persistence/pg/tenant_repo_pg.go
git commit -m "feat(tenant): add PostgreSQL repository implementation"
```

---

## Task 11: Router + Wiring

**Files:**
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Add Tenant to router Handlers struct and routes**

In `backend/internal/delivery/http/router/router.go`:

1. Add `Tenant *handler.TenantHandler` to the `Handlers` struct
2. Add tenant route group inside the `/api/v1` block, after beneficiaries:

```go
// Tenants (authenticated, landlord/admin)
r.Route("/tenants", func(r chi.Router) {
	r.Use(requireAuth)
	r.Use(middleware.RequireRole("admin", "landlord"))
	r.Get("/", h.Tenant.List)
	r.Get("/{id}", h.Tenant.Get)
	r.Post("/", h.Tenant.Create)
	r.Put("/{id}", h.Tenant.Update)
	r.Put("/{id}/deactivate", h.Tenant.Deactivate)
	r.Delete("/{id}", h.Tenant.Delete)
})
```

- [ ] **Step 2: Wire tenant module in main.go**

In `backend/cmd/server/main.go`, after the beneficiary wiring block:

1. Add import: `tenantuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/tenant"`
2. Add wiring:

```go
// Wire Tenant module
tenantRepo := pg.NewTenantRepoPG(pgPool)
createTenantUC := tenantuc.NewCreateTenantUseCase(tenantRepo, userRepo)
getTenantUC := tenantuc.NewGetTenantUseCase(tenantRepo)
listTenantsUC := tenantuc.NewListTenantsUseCase(tenantRepo)
updateTenantUC := tenantuc.NewUpdateTenantUseCase(tenantRepo)
deactivateTenantUC := tenantuc.NewDeactivateTenantUseCase(tenantRepo)
deleteTenantUC := tenantuc.NewDeleteTenantUseCase(tenantRepo)
tenantHandler := handler.NewTenantHandler(createTenantUC, getTenantUC, listTenantsUC, updateTenantUC, deactivateTenantUC, deleteTenantUC)
```

3. Add `Tenant: tenantHandler,` to the `router.Handlers{}` struct literal

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go build ./cmd/server/`
Expected: No errors

- [ ] **Step 4: Run all tests**

Run: `cd backend && go test ./tests/... -v`
Expected: All tests pass (previous 88 + ~38 new = ~126 total)

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/router/router.go backend/cmd/server/main.go
git commit -m "feat(tenant): wire tenant module into router and main.go"
```

---

## Task 12: Final Verification

- [ ] **Step 1: Run full test suite**

Run: `cd backend && go test ./tests/... -count=1 -v`
Expected: All tests pass

- [ ] **Step 2: Build the binary**

Run: `cd backend && go build -o /dev/null ./cmd/server/`
Expected: Builds successfully

- [ ] **Step 3: Update CLAUDE.md**

Add Tenants to the API Routes table and Current Modules section.

- [ ] **Step 4: Final commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with Tenants module"
```
