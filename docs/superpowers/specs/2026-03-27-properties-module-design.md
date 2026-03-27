# Properties Module Design Spec

**Date:** 2026-03-27
**Status:** Approved
**Approach:** Faithful Port (same as Tenants) — port Property entity as-is from business logic reference.

## Context

The Properties module represents land parcels and buildings owned by landlords. Properties can be linked to debts (future module). Each property belongs to exactly one owner (User with landlord/admin role). Reuses the Address value object created in the Tenants module.

**Completed modules:** Programs, Users & Auth, Beneficiaries, Tenants
**Build order:** ~~Tenants~~ -> Properties -> Debts -> Transactions -> Audit -> Dashboard

## 1. Domain Entity

### Property Entity (`domain/entity/property.go`)

```go
type Property struct {
    ID                  uuid.UUID
    Name                string      // required
    PropertyCode        string      // unique, required
    Address             *Address    // nullable, reuses Address value object
    GeoJSONCoordinates  *string     // nullable
    PropertyType        string      // default "LAND"
    SizeInAcres         *float64    // nullable
    SizeInSqm           float64     // required, must be > 0
    OwnerID             uuid.UUID   // FK -> User
    IsAvailableForRent  bool        // default true
    IsActive            bool        // default true
    MonthlyRentAmount   *float64    // nullable
    Description         *string     // nullable
    CreatedAt           time.Time
    UpdatedAt           time.Time
    DeletedAt           *time.Time  // soft delete
}
```

**Validation rules** (in `Validate()` method):
- `Name` required, non-empty, max 255 chars
- `PropertyCode` required, non-empty
- `SizeInSqm` must be > 0
- `PropertyType` defaults to "LAND" if empty

Domain errors:
- `ErrPropertyNameRequired`
- `ErrPropertyNameTooLong`
- `ErrPropertyCodeRequired`
- `ErrPropertySizeRequired` (SizeInSqm <= 0)
- `ErrPropertyCodeExists`

## 2. Repository Interface

### `domain/repository/property_repository.go`

```go
type PropertyFilter struct {
    OwnerID            *uuid.UUID
    IsActive           *bool
    IsAvailableForRent *bool
    PropertyType       *string
    Search             *string  // searches name, property_code
    Limit              int
    Offset             int
}

type PropertyRepository interface {
    Create(ctx context.Context, property *entity.Property) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Property, error)
    GetByPropertyCode(ctx context.Context, code string) (*entity.Property, error)
    List(ctx context.Context, filter PropertyFilter) ([]*entity.Property, int, error)
    Update(ctx context.Context, property *entity.Property) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

## 3. Use Cases

Six use cases in `usecase/property/`:

| Use Case | Key Logic |
|:---|:---|
| CreateProperty | Validates owner exists, checks property_code uniqueness |
| GetProperty | Gets by ID, returns not found if missing |
| ListProperties | Filters by owner, active status, type, rent availability, pagination |
| UpdateProperty | Existence check, re-validates, preserves immutable fields |
| DeactivateProperty | Sets is_active=false (debt check deferred to Debts module) |
| DeleteProperty | Soft delete via DeletedAt |

## 4. Database Migration

### `000008_create_properties.up.sql`

- Properties table with UUID PK, address as flat columns (same pattern as tenants)
- owner_id FK to users(id) ON DELETE RESTRICT
- Unique partial index on property_code WHERE deleted_at IS NULL
- Indexes: owner_id, is_active, property_type, name (gin_trgm_ops)
- Auto-updated_at trigger

## 5. DTOs

- `CreatePropertyRequest` — name, property_code, address (AddressDTO), geojson_coordinates, property_type, size_in_acres, size_in_sqm, monthly_rent_amount, description
- `UpdatePropertyRequest` — same fields, all updatable
- `PropertyResponse` — all fields with nested AddressDTO, timestamps
- `PropertyListResponse` — array + total/limit/offset

## 6. HTTP Handler & Routes

Standard CRUD + Deactivate with Swagger annotations. All routes under `/api/v1/properties`, auth + landlord/admin role required.

## 7. Tests

| File | Estimated Count |
|:---|:---|
| `property_entity_test.go` | ~8 tests |
| `property_usecase_test.go` | ~16 tests |
| `property_handler_test.go` | ~11 tests |

~35 new tests, bringing total to ~157.

## Files to Create/Modify

### New Files
1. `backend/internal/domain/entity/property.go`
2. `backend/internal/domain/repository/property_repository.go`
3. `backend/internal/usecase/property/create_property.go`
4. `backend/internal/usecase/property/get_property.go`
5. `backend/internal/usecase/property/list_properties.go`
6. `backend/internal/usecase/property/update_property.go`
7. `backend/internal/usecase/property/deactivate_property.go`
8. `backend/internal/usecase/property/delete_property.go`
9. `backend/internal/delivery/http/dto/property_dto.go` (NOTE: this file already exists for Programs — tenant module uses a SEPARATE file. Need to use a new name like `property_land_dto.go` or extend the existing file. Actually the existing `property_dto.go` is for Programs. We should name this `land_property_dto.go` or just add to a new file.)
10. `backend/internal/delivery/http/handler/property_handler.go`
11. `backend/internal/infrastructure/persistence/pg/property_repo_pg.go`
12. `backend/migrations/000008_create_properties.up.sql`
13. `backend/migrations/000008_create_properties.down.sql`
14. `backend/tests/mocks/property_repository_mock.go`
15. `backend/tests/unit/property_land_entity_test.go`
16. `backend/tests/unit/property_land_usecase_test.go`
17. `backend/tests/unit/property_land_handler_test.go`

### Modified Files
1. `backend/internal/domain/entity/errors.go` — add property domain errors
2. `backend/internal/delivery/http/router/router.go` — add PropertyHandler + routes
3. `backend/cmd/server/main.go` — wire property module

## Naming Note

The existing `program_dto.go` and `program_handler.go` files are for the Programs domain (social programs). This new Properties module is for land/building properties — entirely different domain. To avoid confusion:
- Entity: `property.go` (the entity name is `Property` — distinct from `Program`)
- DTO: `land_property_dto.go` to distinguish from the program-related file
- Handler: `property_handler.go` (no conflict since `program_handler.go` is the Programs one)
- Tests: prefixed with `property_land_` to distinguish

## Deferred
- Debt check on deactivation (will be added when Debts module is built)
- GeoJSON validation (not in original system's Go port scope)
