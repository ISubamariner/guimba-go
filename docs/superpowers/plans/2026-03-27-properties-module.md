# Properties Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the Properties domain module — entity, repository, use cases, persistence, DTOs, handler, routes, and tests — following established Clean Architecture patterns.

**Architecture:** Faithful port of Property entity from business logic reference. Properties belong to an owner (User with landlord/admin role). Reuses Address value object from Tenants module. Deactivation implemented without debt check (deferred to Debts module).

**Tech Stack:** Go 1.26+, Chi v5, pgx v5, go-playground/validator v10, google/uuid

**Spec:** `docs/superpowers/specs/2026-03-27-properties-module-design.md`

---

## File Map

| Action | File | Responsibility |
|:---|:---|:---|
| Create | `backend/internal/domain/entity/property.go` | Property entity + validation |
| Modify | `backend/internal/domain/entity/errors.go` | Add property domain errors |
| Create | `backend/internal/domain/repository/property_repository.go` | PropertyRepository interface + PropertyFilter |
| Create | `backend/tests/mocks/property_repository_mock.go` | Manual mock |
| Create | `backend/internal/usecase/property/create_property.go` | CreateProperty use case |
| Create | `backend/internal/usecase/property/get_property.go` | GetProperty use case |
| Create | `backend/internal/usecase/property/list_properties.go` | ListProperties use case |
| Create | `backend/internal/usecase/property/update_property.go` | UpdateProperty use case |
| Create | `backend/internal/usecase/property/deactivate_property.go` | DeactivateProperty use case |
| Create | `backend/internal/usecase/property/delete_property.go` | DeleteProperty use case |
| Create | `backend/internal/delivery/http/dto/property_dto.go` | Request/response DTOs + converters |
| Create | `backend/internal/delivery/http/handler/property_handler.go` | HTTP handlers with Swagger |
| Create | `backend/internal/infrastructure/persistence/pg/property_repo_pg.go` | PostgreSQL repository |
| Create | `backend/migrations/000008_create_properties.up.sql` | Properties table |
| Create | `backend/migrations/000008_create_properties.down.sql` | Drop properties table |
| Modify | `backend/internal/delivery/http/router/router.go` | Add Property field + routes |
| Modify | `backend/cmd/server/main.go` | Wire property module |
| Create | `backend/tests/unit/property_entity_test.go` | Entity validation tests |
| Create | `backend/tests/unit/property_usecase_test.go` | Use case tests |
| Create | `backend/tests/unit/property_handler_test.go` | Handler tests |

---

## Task 1: Domain Layer (Entity + Errors + Repo Interface + Mock)

**Files:**
- Create: `backend/internal/domain/entity/property.go`
- Modify: `backend/internal/domain/entity/errors.go`
- Create: `backend/internal/domain/repository/property_repository.go`
- Create: `backend/tests/mocks/property_repository_mock.go`
- Create: `backend/tests/unit/property_entity_test.go`

- [ ] **Step 1: Add property domain errors to errors.go**

Append to `backend/internal/domain/entity/errors.go`:

```go
// Domain errors for Property entity.
var (
	ErrPropertyNameRequired = errors.New("property name is required")
	ErrPropertyNameTooLong  = errors.New("property name must be 255 characters or less")
	ErrPropertyCodeRequired = errors.New("property code is required")
	ErrPropertySizeRequired = errors.New("property size in square meters must be greater than zero")
	ErrPropertyCodeExists   = errors.New("a property with this code already exists")
)
```

- [ ] **Step 2: Create Property entity**

Create `backend/internal/domain/entity/property.go`:

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Property represents a land parcel or building owned by a landlord.
type Property struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	PropertyCode       string     `json:"property_code"`
	Address            *Address   `json:"address,omitempty"`
	GeoJSONCoordinates *string    `json:"geojson_coordinates,omitempty"`
	PropertyType       string     `json:"property_type"`
	SizeInAcres        *float64   `json:"size_in_acres,omitempty"`
	SizeInSqm          float64    `json:"size_in_sqm"`
	OwnerID            uuid.UUID  `json:"owner_id"`
	IsAvailableForRent bool       `json:"is_available_for_rent"`
	IsActive           bool       `json:"is_active"`
	MonthlyRentAmount  *float64   `json:"monthly_rent_amount,omitempty"`
	Description        *string    `json:"description,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// NewProperty creates a new Property with generated ID and defaults.
func NewProperty(name, propertyCode string, address *Address, geojson *string, propertyType string, sizeInAcres *float64, sizeInSqm float64, ownerID uuid.UUID, monthlyRent *float64, description *string) (*Property, error) {
	if propertyType == "" {
		propertyType = "LAND"
	}

	p := &Property{
		ID:                 uuid.New(),
		Name:               name,
		PropertyCode:       propertyCode,
		Address:            address,
		GeoJSONCoordinates: geojson,
		PropertyType:       propertyType,
		SizeInAcres:        sizeInAcres,
		SizeInSqm:          sizeInSqm,
		OwnerID:            ownerID,
		IsAvailableForRent: true,
		IsActive:           true,
		MonthlyRentAmount:  monthlyRent,
		Description:        description,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

// Validate checks business rules for a Property.
func (p *Property) Validate() error {
	if p.Name == "" {
		return ErrPropertyNameRequired
	}
	if len(p.Name) > 255 {
		return ErrPropertyNameTooLong
	}
	if p.PropertyCode == "" {
		return ErrPropertyCodeRequired
	}
	if p.SizeInSqm <= 0 {
		return ErrPropertySizeRequired
	}
	return nil
}
```

- [ ] **Step 3: Create repository interface**

Create `backend/internal/domain/repository/property_repository.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// PropertyFilter holds optional filters for listing properties.
type PropertyFilter struct {
	OwnerID            *uuid.UUID
	IsActive           *bool
	IsAvailableForRent *bool
	PropertyType       *string
	Search             *string
	Limit              int
	Offset             int
}

// PropertyRepository defines the interface for property persistence operations.
type PropertyRepository interface {
	Create(ctx context.Context, property *entity.Property) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Property, error)
	GetByPropertyCode(ctx context.Context, code string) (*entity.Property, error)
	List(ctx context.Context, filter PropertyFilter) ([]*entity.Property, int, error)
	Update(ctx context.Context, property *entity.Property) error
	Delete(ctx context.Context, id uuid.UUID) error
}
```

- [ ] **Step 4: Create mock**

Create `backend/tests/mocks/property_repository_mock.go`:

```go
package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// PropertyRepositoryMock is a test mock for repository.PropertyRepository.
type PropertyRepositoryMock struct {
	CreateFn            func(ctx context.Context, property *entity.Property) error
	GetByIDFn           func(ctx context.Context, id uuid.UUID) (*entity.Property, error)
	GetByPropertyCodeFn func(ctx context.Context, code string) (*entity.Property, error)
	ListFn              func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error)
	UpdateFn            func(ctx context.Context, property *entity.Property) error
	DeleteFn            func(ctx context.Context, id uuid.UUID) error
}

func (m *PropertyRepositoryMock) Create(ctx context.Context, property *entity.Property) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, property)
	}
	return nil
}

func (m *PropertyRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *PropertyRepositoryMock) GetByPropertyCode(ctx context.Context, code string) (*entity.Property, error) {
	if m.GetByPropertyCodeFn != nil {
		return m.GetByPropertyCodeFn(ctx, code)
	}
	return nil, nil
}

func (m *PropertyRepositoryMock) List(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *PropertyRepositoryMock) Update(ctx context.Context, property *entity.Property) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, property)
	}
	return nil
}

func (m *PropertyRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
```

- [ ] **Step 5: Write entity tests**

Create `backend/tests/unit/property_entity_test.go`:

```go
package unit

import (
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/google/uuid"
)

func TestNewProperty_Valid(t *testing.T) {
	p, err := entity.NewProperty("Farm Plot A", "FP-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name != "Farm Plot A" {
		t.Errorf("expected name 'Farm Plot A', got %q", p.Name)
	}
	if p.PropertyCode != "FP-001" {
		t.Errorf("expected code 'FP-001', got %q", p.PropertyCode)
	}
	if p.PropertyType != "LAND" {
		t.Errorf("expected type 'LAND', got %q", p.PropertyType)
	}
	if !p.IsActive {
		t.Error("expected IsActive true")
	}
	if !p.IsAvailableForRent {
		t.Error("expected IsAvailableForRent true")
	}
	if p.ID == (uuid.UUID{}) {
		t.Error("expected non-zero UUID")
	}
}

func TestNewProperty_DefaultType(t *testing.T) {
	p, err := entity.NewProperty("Plot", "P-001", nil, nil, "", nil, 100.0, uuid.New(), nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.PropertyType != "LAND" {
		t.Errorf("expected default type 'LAND', got %q", p.PropertyType)
	}
}

func TestNewProperty_NameRequired(t *testing.T) {
	_, err := entity.NewProperty("", "P-001", nil, nil, "LAND", nil, 100.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertyNameRequired {
		t.Errorf("expected ErrPropertyNameRequired, got %v", err)
	}
}

func TestNewProperty_NameTooLong(t *testing.T) {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := entity.NewProperty(string(longName), "P-001", nil, nil, "LAND", nil, 100.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertyNameTooLong {
		t.Errorf("expected ErrPropertyNameTooLong, got %v", err)
	}
}

func TestNewProperty_CodeRequired(t *testing.T) {
	_, err := entity.NewProperty("Plot", "", nil, nil, "LAND", nil, 100.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertyCodeRequired {
		t.Errorf("expected ErrPropertyCodeRequired, got %v", err)
	}
}

func TestNewProperty_SizeRequired(t *testing.T) {
	_, err := entity.NewProperty("Plot", "P-001", nil, nil, "LAND", nil, 0, uuid.New(), nil, nil)
	if err != entity.ErrPropertySizeRequired {
		t.Errorf("expected ErrPropertySizeRequired, got %v", err)
	}
}

func TestNewProperty_NegativeSize(t *testing.T) {
	_, err := entity.NewProperty("Plot", "P-001", nil, nil, "LAND", nil, -5.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertySizeRequired {
		t.Errorf("expected ErrPropertySizeRequired, got %v", err)
	}
}

func TestNewProperty_WithAllFields(t *testing.T) {
	addr := entity.NewAddress("123 Farm Rd", "Guimba", "Nueva Ecija", "3115", "Philippines")
	geojson := `{"type":"Point","coordinates":[120.77,15.66]}`
	acres := 1.24
	rent := 5000.0
	desc := "Rice paddy"

	p, err := entity.NewProperty("Big Farm", "BF-001", addr, &geojson, "AGRICULTURAL", &acres, 5000.0, uuid.New(), &rent, &desc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Address == nil || p.Address.City != "Guimba" {
		t.Error("address mismatch")
	}
	if *p.GeoJSONCoordinates != geojson {
		t.Error("geojson mismatch")
	}
	if *p.SizeInAcres != acres {
		t.Error("acres mismatch")
	}
	if *p.MonthlyRentAmount != rent {
		t.Error("rent mismatch")
	}
	if *p.Description != desc {
		t.Error("description mismatch")
	}
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend && go test ./tests/unit/ -run "TestNewProperty" -v`
Expected: All 8 tests PASS

- [ ] **Step 7: Verify builds**

Run: `cd backend && go build ./internal/domain/... && go build ./tests/mocks/`
Expected: No errors

- [ ] **Step 8: Commit**

```bash
git add backend/internal/domain/entity/property.go backend/internal/domain/entity/errors.go backend/internal/domain/repository/property_repository.go backend/tests/mocks/property_repository_mock.go backend/tests/unit/property_entity_test.go
git commit -m "feat(property): add Property entity, repo interface, mock, and entity tests"
```

---

## Task 2: Use Cases + Tests

**Files:**
- Create: `backend/internal/usecase/property/create_property.go`
- Create: `backend/internal/usecase/property/get_property.go`
- Create: `backend/internal/usecase/property/list_properties.go`
- Create: `backend/internal/usecase/property/update_property.go`
- Create: `backend/internal/usecase/property/deactivate_property.go`
- Create: `backend/internal/usecase/property/delete_property.go`
- Create: `backend/tests/unit/property_usecase_test.go`

- [ ] **Step 1: Create all 6 use cases**

`backend/internal/usecase/property/create_property.go`:
```go
package property

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreatePropertyUseCase struct {
	repo     repository.PropertyRepository
	userRepo repository.UserRepository
}

func NewCreatePropertyUseCase(repo repository.PropertyRepository, userRepo repository.UserRepository) *CreatePropertyUseCase {
	return &CreatePropertyUseCase{repo: repo, userRepo: userRepo}
}

func (uc *CreatePropertyUseCase) Execute(ctx context.Context, property *entity.Property) error {
	if err := property.Validate(); err != nil {
		return err
	}

	owner, err := uc.userRepo.GetByID(ctx, property.OwnerID)
	if err != nil {
		return err
	}
	if owner == nil {
		return apperror.NewNotFound("User", property.OwnerID)
	}

	existing, err := uc.repo.GetByPropertyCode(ctx, property.PropertyCode)
	if err != nil {
		return err
	}
	if existing != nil {
		return apperror.NewConflict(entity.ErrPropertyCodeExists.Error())
	}

	return uc.repo.Create(ctx, property)
}
```

`backend/internal/usecase/property/get_property.go`:
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type GetPropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewGetPropertyUseCase(repo repository.PropertyRepository) *GetPropertyUseCase {
	return &GetPropertyUseCase{repo: repo}
}

func (uc *GetPropertyUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
	property, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, apperror.NewNotFound("Property", id)
	}
	return property, nil
}
```

`backend/internal/usecase/property/list_properties.go`:
```go
package property

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListPropertiesUseCase struct {
	repo repository.PropertyRepository
}

func NewListPropertiesUseCase(repo repository.PropertyRepository) *ListPropertiesUseCase {
	return &ListPropertiesUseCase{repo: repo}
}

func (uc *ListPropertiesUseCase) Execute(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
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

`backend/internal/usecase/property/update_property.go`:
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdatePropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewUpdatePropertyUseCase(repo repository.PropertyRepository) *UpdatePropertyUseCase {
	return &UpdatePropertyUseCase{repo: repo}
}

func (uc *UpdatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID, property *entity.Property) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	property.ID = id
	property.CreatedAt = existing.CreatedAt
	property.OwnerID = existing.OwnerID

	if err := property.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, property)
}
```

`backend/internal/usecase/property/deactivate_property.go`:
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivatePropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewDeactivatePropertyUseCase(repo repository.PropertyRepository) *DeactivatePropertyUseCase {
	return &DeactivatePropertyUseCase{repo: repo}
}

func (uc *DeactivatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	existing.IsActive = false
	return uc.repo.Update(ctx, existing)
}
```

`backend/internal/usecase/property/delete_property.go`:
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeletePropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewDeletePropertyUseCase(repo repository.PropertyRepository) *DeletePropertyUseCase {
	return &DeletePropertyUseCase{repo: repo}
}

func (uc *DeletePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	return uc.repo.Delete(ctx, id)
}
```

- [ ] **Step 2: Write use case tests**

Create `backend/tests/unit/property_usecase_test.go`:

```go
package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	property "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- CreateProperty ---

func TestCreateProperty_Success(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByPropertyCodeFn: func(ctx context.Context, code string) (*entity.Property, error) {
			return nil, nil
		},
		CreateFn: func(ctx context.Context, p *entity.Property) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	uc := property.NewCreatePropertyUseCase(repo, userRepo)
	p, _ := entity.NewProperty("Farm", "F-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	err := uc.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateProperty_OwnerNotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, nil
		},
	}

	uc := property.NewCreatePropertyUseCase(repo, userRepo)
	p, _ := entity.NewProperty("Farm", "F-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	err := uc.Execute(context.Background(), p)
	if err == nil {
		t.Fatal("expected error when owner not found")
	}
}

func TestCreateProperty_DuplicateCode(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByPropertyCodeFn: func(ctx context.Context, code string) (*entity.Property, error) {
			return &entity.Property{ID: uuid.New(), PropertyCode: code}, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	uc := property.NewCreatePropertyUseCase(repo, userRepo)
	p, _ := entity.NewProperty("Farm", "F-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	err := uc.Execute(context.Background(), p)
	if err == nil {
		t.Fatal("expected error for duplicate code")
	}
}

// --- GetProperty ---

func TestGetProperty_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, OwnerID: uuid.New()}, nil
		},
	}
	uc := property.NewGetPropertyUseCase(repo)
	result, err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != propID {
		t.Error("expected ID to match")
	}
}

func TestGetProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewGetPropertyUseCase(repo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- ListProperties ---

func TestListProperties_Success(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			return []*entity.Property{{ID: uuid.New(), Name: "Farm", SizeInSqm: 500}}, 1, nil
		},
	}
	uc := property.NewListPropertiesUseCase(repo)
	props, total, err := uc.Execute(context.Background(), repository.PropertyFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 || len(props) != 1 {
		t.Errorf("expected 1 property, got %d", len(props))
	}
}

func TestListProperties_DefaultLimit(t *testing.T) {
	var captured repository.PropertyFilter
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := property.NewListPropertiesUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.PropertyFilter{Limit: 0})
	if captured.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", captured.Limit)
	}
}

func TestListProperties_MaxLimit(t *testing.T) {
	var captured repository.PropertyFilter
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := property.NewListPropertiesUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.PropertyFilter{Limit: 500})
	if captured.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", captured.Limit)
	}
}

// --- UpdateProperty ---

func TestUpdateProperty_Success(t *testing.T) {
	propID := uuid.New()
	ownerID := uuid.New()
	existing := &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, OwnerID: ownerID, IsActive: true}

	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Property) error { return nil },
	}

	uc := property.NewUpdatePropertyUseCase(repo)
	updated := &entity.Property{Name: "Updated Farm", PropertyCode: "F-001", SizeInSqm: 600}
	err := uc.Execute(context.Background(), propID, updated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.OwnerID != ownerID {
		t.Error("expected OwnerID to be preserved")
	}
}

func TestUpdateProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewUpdatePropertyUseCase(repo)
	err := uc.Execute(context.Background(), uuid.New(), &entity.Property{Name: "X", PropertyCode: "X", SizeInSqm: 100})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- DeactivateProperty ---

func TestDeactivateProperty_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, IsActive: true}, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Property) error {
			if p.IsActive {
				t.Error("expected IsActive to be false")
			}
			return nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(repo)
	err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeactivateProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(repo)
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- DeleteProperty ---

func TestDeleteProperty_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	uc := property.NewDeletePropertyUseCase(repo)
	err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewDeletePropertyUseCase(repo)
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}
```

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./tests/unit/ -run "TestCreateProperty|TestGetProperty|TestListProperties|TestUpdateProperty|TestDeactivateProperty|TestDeleteProperty" -v`
Expected: All 14 tests PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/usecase/property/ backend/tests/unit/property_usecase_test.go
git commit -m "feat(property): add all 6 use cases with tests"
```

---

## Task 3: DTOs + Handler + Handler Tests

**Files:**
- Create: `backend/internal/delivery/http/dto/property_dto.go`
- Create: `backend/internal/delivery/http/handler/property_handler.go`
- Create: `backend/tests/unit/property_handler_test.go`

- [ ] **Step 1: Create property DTOs**

Create `backend/internal/delivery/http/dto/property_dto.go`:

```go
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// CreatePropertyRequest is the request body for creating a property.
type CreatePropertyRequest struct {
	Name               string      `json:"name" validate:"required,max=255"`
	PropertyCode       string      `json:"property_code" validate:"required,max=100"`
	Address            *AddressDTO `json:"address" validate:"omitempty"`
	GeoJSONCoordinates *string     `json:"geojson_coordinates" validate:"omitempty"`
	PropertyType       string      `json:"property_type" validate:"omitempty,max=50"`
	SizeInAcres        *float64    `json:"size_in_acres" validate:"omitempty,gt=0"`
	SizeInSqm          float64     `json:"size_in_sqm" validate:"required,gt=0"`
	MonthlyRentAmount  *float64    `json:"monthly_rent_amount" validate:"omitempty,gte=0"`
	Description        *string     `json:"description" validate:"omitempty"`
}

// UpdatePropertyRequest is the request body for updating a property.
type UpdatePropertyRequest struct {
	Name               string      `json:"name" validate:"required,max=255"`
	PropertyCode       string      `json:"property_code" validate:"required,max=100"`
	Address            *AddressDTO `json:"address" validate:"omitempty"`
	GeoJSONCoordinates *string     `json:"geojson_coordinates" validate:"omitempty"`
	PropertyType       string      `json:"property_type" validate:"omitempty,max=50"`
	SizeInAcres        *float64    `json:"size_in_acres" validate:"omitempty,gt=0"`
	SizeInSqm          float64     `json:"size_in_sqm" validate:"required,gt=0"`
	IsAvailableForRent *bool       `json:"is_available_for_rent" validate:"omitempty"`
	MonthlyRentAmount  *float64    `json:"monthly_rent_amount" validate:"omitempty,gte=0"`
	Description        *string     `json:"description" validate:"omitempty"`
}

// PropertyResponse is the response body for a single property.
type PropertyResponse struct {
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	PropertyCode       string      `json:"property_code"`
	Address            *AddressDTO `json:"address,omitempty"`
	GeoJSONCoordinates *string     `json:"geojson_coordinates,omitempty"`
	PropertyType       string      `json:"property_type"`
	SizeInAcres        *float64    `json:"size_in_acres,omitempty"`
	SizeInSqm          float64     `json:"size_in_sqm"`
	OwnerID            uuid.UUID   `json:"owner_id"`
	IsAvailableForRent bool        `json:"is_available_for_rent"`
	IsActive           bool        `json:"is_active"`
	MonthlyRentAmount  *float64    `json:"monthly_rent_amount,omitempty"`
	Description        *string     `json:"description,omitempty"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
}

// PropertyListResponse is the response body for a list of properties.
type PropertyListResponse struct {
	Data   []PropertyResponse `json:"data"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// ToEntity converts a CreatePropertyRequest to a domain entity.
func (r *CreatePropertyRequest) ToEntity(ownerID uuid.UUID) (*entity.Property, error) {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	return entity.NewProperty(r.Name, r.PropertyCode, addr, r.GeoJSONCoordinates, r.PropertyType, r.SizeInAcres, r.SizeInSqm, ownerID, r.MonthlyRentAmount, r.Description)
}

// ToEntity converts an UpdatePropertyRequest to a partial domain entity.
func (r *UpdatePropertyRequest) ToEntity() *entity.Property {
	var addr *entity.Address
	if r.Address != nil {
		addr = entity.NewAddress(r.Address.Street, r.Address.City, r.Address.StateOrRegion, r.Address.PostalCode, r.Address.Country)
	}
	p := &entity.Property{
		Name:               r.Name,
		PropertyCode:       r.PropertyCode,
		Address:            addr,
		GeoJSONCoordinates: r.GeoJSONCoordinates,
		PropertyType:       r.PropertyType,
		SizeInAcres:        r.SizeInAcres,
		SizeInSqm:          r.SizeInSqm,
		IsAvailableForRent: true,
		MonthlyRentAmount:  r.MonthlyRentAmount,
		Description:        r.Description,
	}
	if r.IsAvailableForRent != nil {
		p.IsAvailableForRent = *r.IsAvailableForRent
	}
	return p
}

// NewPropertyResponse creates a PropertyResponse from a domain entity.
func NewPropertyResponse(p *entity.Property) PropertyResponse {
	resp := PropertyResponse{
		ID:                 p.ID,
		Name:               p.Name,
		PropertyCode:       p.PropertyCode,
		GeoJSONCoordinates: p.GeoJSONCoordinates,
		PropertyType:       p.PropertyType,
		SizeInAcres:        p.SizeInAcres,
		SizeInSqm:          p.SizeInSqm,
		OwnerID:            p.OwnerID,
		IsAvailableForRent: p.IsAvailableForRent,
		IsActive:           p.IsActive,
		MonthlyRentAmount:  p.MonthlyRentAmount,
		Description:        p.Description,
		CreatedAt:          p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          p.UpdatedAt.Format(time.RFC3339),
	}
	if p.Address != nil {
		resp.Address = &AddressDTO{
			Street:        p.Address.Street,
			City:          p.Address.City,
			StateOrRegion: p.Address.StateOrRegion,
			PostalCode:    p.Address.PostalCode,
			Country:       p.Address.Country,
		}
	}
	return resp
}

// NewPropertyListResponse creates a PropertyListResponse from domain entities.
func NewPropertyListResponse(properties []*entity.Property, total, limit, offset int) PropertyListResponse {
	data := make([]PropertyResponse, 0, len(properties))
	for _, p := range properties {
		data = append(data, NewPropertyResponse(p))
	}
	return PropertyListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
```

- [ ] **Step 2: Create property handler**

Create `backend/internal/delivery/http/handler/property_handler.go` — follows exact same pattern as `tenant_handler.go`. All 6 methods (Create, Get, List, Update, Deactivate, Delete) with Swagger annotations. Uses `handlePropertyDomainError` to map `ErrPropertyNameRequired`, `ErrPropertyNameTooLong`, `ErrPropertyCodeRequired`, `ErrPropertySizeRequired` to 422, and `ErrPropertyCodeExists` to 409. Extracts `ownerID` from auth context for Create.

Handler struct holds: createUC, getUC, listUC, updateUC, deactivateUC, deleteUC. Constructor: `NewPropertyHandler(...)`.

List endpoint supports query params: `owner_id`, `is_active`, `is_available_for_rent`, `property_type`, `search`, `limit`, `offset`.

- [ ] **Step 3: Create handler tests**

Create `backend/tests/unit/property_handler_test.go` — follows same pattern as `tenant_handler_test.go`. Tests: Create_Success, Create_InvalidJSON, Create_ValidationFailed, Get_Success, Get_InvalidID, Get_NotFound, List_Success, Delete_Success (~8 tests).

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./tests/unit/ -run "TestPropertyHandler" -v`
Expected: All handler tests PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/dto/property_dto.go backend/internal/delivery/http/handler/property_handler.go backend/tests/unit/property_handler_test.go
git commit -m "feat(property): add DTOs, HTTP handler with Swagger, and handler tests"
```

---

## Task 4: Migration + PG Repository

**Files:**
- Create: `backend/migrations/000008_create_properties.up.sql`
- Create: `backend/migrations/000008_create_properties.down.sql`
- Create: `backend/internal/infrastructure/persistence/pg/property_repo_pg.go`

- [ ] **Step 1: Create up migration**

```sql
-- Properties table
CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    property_code VARCHAR(100) NOT NULL,
    address_street VARCHAR(255),
    address_city VARCHAR(255),
    address_state_or_region VARCHAR(255),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100) DEFAULT 'Philippines',
    geojson_coordinates TEXT,
    property_type VARCHAR(50) NOT NULL DEFAULT 'LAND',
    size_in_acres DECIMAL(12,4),
    size_in_sqm DECIMAL(12,4) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_available_for_rent BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    monthly_rent_amount DECIMAL(12,2),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Trigger for auto-updating updated_at
CREATE TRIGGER set_properties_updated_at
    BEFORE UPDATE ON properties
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Indexes
CREATE UNIQUE INDEX idx_properties_property_code ON properties (property_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_owner_id ON properties (owner_id);
CREATE INDEX idx_properties_is_active ON properties (is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_property_type ON properties (property_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_deleted_at ON properties (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_properties_name ON properties USING gin (name gin_trgm_ops);
```

- [ ] **Step 2: Create down migration**

```sql
DROP TRIGGER IF EXISTS set_properties_updated_at ON properties;
DROP TABLE IF EXISTS properties;
```

- [ ] **Step 3: Create PG repository**

Create `backend/internal/infrastructure/persistence/pg/property_repo_pg.go` — follows same pattern as `tenant_repo_pg.go`. Key differences:
- Scans property_code, geojson_coordinates, property_type, size_in_acres, size_in_sqm, is_available_for_rent, monthly_rent_amount, description fields
- Address mapped to/from flat columns (same as tenants)
- `GetByPropertyCode` method instead of `GetByEmail`
- List filters: owner_id, is_active, is_available_for_rent, property_type, search (ILIKE on name and property_code)
- Helper methods: `scanProperty(row)` and `scanPropertyRow(rows)`

- [ ] **Step 4: Verify builds**

Run: `cd backend && go build ./internal/infrastructure/persistence/pg/`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/000008_create_properties.up.sql backend/migrations/000008_create_properties.down.sql backend/internal/infrastructure/persistence/pg/property_repo_pg.go
git commit -m "feat(property): add database migration and PostgreSQL repository"
```

---

## Task 5: Router + Wiring + Final Verification

**Files:**
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add Property to router**

In `router.go`:
1. Add `Property *handler.PropertyHandler` to `Handlers` struct
2. Add route group after tenants:

```go
// Properties (authenticated, landlord/admin)
r.Route("/properties", func(r chi.Router) {
    r.Use(requireAuth)
    r.Use(middleware.RequireRole("admin", "landlord"))
    r.Get("/", h.Property.List)
    r.Get("/{id}", h.Property.Get)
    r.Post("/", h.Property.Create)
    r.Put("/{id}", h.Property.Update)
    r.Put("/{id}/deactivate", h.Property.Deactivate)
    r.Delete("/{id}", h.Property.Delete)
})
```

- [ ] **Step 2: Wire in main.go**

Add import: `propertyuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"`

Add wiring block after tenants:
```go
// Wire Property module
propertyRepo := pg.NewPropertyRepoPG(pgPool)
createPropertyUC := propertyuc.NewCreatePropertyUseCase(propertyRepo, userRepo)
getPropertyUC := propertyuc.NewGetPropertyUseCase(propertyRepo)
listPropertiesUC := propertyuc.NewListPropertiesUseCase(propertyRepo)
updatePropertyUC := propertyuc.NewUpdatePropertyUseCase(propertyRepo)
deactivatePropertyUC := propertyuc.NewDeactivatePropertyUseCase(propertyRepo)
deletePropertyUC := propertyuc.NewDeletePropertyUseCase(propertyRepo)
propertyHandler := handler.NewPropertyHandler(createPropertyUC, getPropertyUC, listPropertiesUC, updatePropertyUC, deactivatePropertyUC, deletePropertyUC)
```

Add `Property: propertyHandler,` to `router.Handlers{}` struct.

- [ ] **Step 3: Build and test**

Run: `cd backend && go build ./cmd/server/ && go test ./tests/... -count=1 -v`
Expected: Build succeeds, all tests pass

- [ ] **Step 4: Update CLAUDE.md**

Add Properties to API Routes table and Current Modules.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/router/router.go backend/cmd/server/main.go CLAUDE.md
git commit -m "feat(property): wire property module and update docs"
```
