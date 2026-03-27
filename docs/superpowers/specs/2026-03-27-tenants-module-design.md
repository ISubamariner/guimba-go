# Tenants Module Design Spec

**Date:** 2026-03-27
**Status:** Approved
**Approach:** Faithful Port (Approach A) — port Tenant entity as-is from business logic reference, defer user account linking to later iteration.

## Context

The Tenants module is the first of the core debt-tracking modules. Tenants are people who owe money to landlords. Each tenant belongs to exactly one landlord, enabling multi-tenant data isolation. This module unblocks Properties, Debts, and Transactions.

**Completed modules:** Programs, Users & Auth (JWT + RBAC), Beneficiaries
**Build order:** Tenants -> Properties -> Debts -> Transactions -> Audit -> Dashboard

## 1. Domain Entity & Value Object

### Address Value Object (`domain/entity/address.go`)

Reusable by Properties module later.

```go
type Address struct {
    Street        string
    City          string
    StateOrRegion string
    PostalCode    string  // nullable
    Country       string  // default "Philippines"
}
```

`FullAddress()` method returns formatted string: "street, city, state, postal_code, country".

### Tenant Entity (`domain/entity/tenant.go`)

```go
type Tenant struct {
    ID          uuid.UUID
    FullName    string
    Email       *string       // nullable
    PhoneNumber *string       // nullable
    NationalID  *string       // nullable
    Address     *Address      // nullable, embedded
    LandlordID  uuid.UUID     // FK -> User
    UserID      *uuid.UUID    // nullable, deferred (not wired in this iteration)
    IsActive    bool          // default true
    Notes       *string       // nullable
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time    // soft delete
}
```

**Validation rules** (in `Validate()` method):
- `FullName` required, non-empty
- Must have at least one contact method (Email OR PhoneNumber)
- If Email provided, must be valid format

Domain errors added to `domain/entity/errors.go`:
- `ErrTenantNotFound`
- `ErrTenantEmailExists`
- `ErrTenantContactRequired` (no email or phone)
- `ErrTenantInvalidEmail`

## 2. Repository Interface

### `domain/repository/tenant_repository.go`

```go
type TenantFilter struct {
    LandlordID *uuid.UUID  // required for landlord-scoped queries
    IsActive   *bool
    Search     *string     // searches full_name, email, phone
    Limit      int
    Offset     int
}

type TenantRepository interface {
    Create(ctx context.Context, tenant *entity.Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
    GetByEmail(ctx context.Context, email string) (*entity.Tenant, error)
    List(ctx context.Context, filter TenantFilter) ([]*entity.Tenant, int, error)
    Update(ctx context.Context, tenant *entity.Tenant) error
    Delete(ctx context.Context, id uuid.UUID) error  // soft delete
}
```

Landlord scoping enforced via `TenantFilter.LandlordID` on list queries. Ownership validated at the use case layer for single-entity operations.

## 3. Use Cases

Six use cases in `usecase/tenant/`, one file per use case:

| Use Case | File | Auth | Key Logic |
|:---|:---|:---|:---|
| CreateTenant | `create_tenant.go` | landlord/admin | Validates landlord exists & has LANDLORD+ role, checks email uniqueness, enforces contact method rule |
| GetTenant | `get_tenant.go` | landlord/admin | Gets by ID, validates requesting user owns this tenant (landlord_id match) or is admin |
| ListTenants | `list_tenants.go` | landlord/admin | Filters by landlord_id of requesting user (admins see all), pagination |
| UpdateTenant | `update_tenant.go` | landlord/admin | Ownership check, re-validates contact method rule, email uniqueness if changed |
| DeactivateTenant | `deactivate_tenant.go` | landlord/admin | Ownership check, sets is_active=false |
| DeleteTenant | `delete_tenant.go` | landlord/admin | Ownership check, soft delete via DeletedAt timestamp |

**Ownership check pattern:** Every mutation verifies `tenant.LandlordID == requestingUserID` unless the requesting user has admin role. This enforces multi-tenant data isolation.

## 4. Database Migration

### `000007_create_tenants.up.sql`

```sql
CREATE TABLE tenants (
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
```

**Indexes:**
- `idx_tenants_landlord_id` on `landlord_id`
- `idx_tenants_email` unique partial where `email IS NOT NULL AND deleted_at IS NULL`
- `idx_tenants_is_active` on `is_active`
- `idx_tenants_deleted_at` on `deleted_at` (for soft delete filtering)

Address stored as individual columns (not JSONB) to keep it queryable and consistent with the existing codebase patterns.

Reversible `.down.sql` drops the table.

## 5. DTOs

### `delivery/http/dto/tenant_dto.go`

**AddressDTO** (shared, reusable by Properties):
```go
type AddressDTO struct {
    Street        string `json:"street"`
    City          string `json:"city"`
    StateOrRegion string `json:"state_or_region"`
    PostalCode    string `json:"postal_code,omitempty"`
    Country       string `json:"country,omitempty"`
}
```

**Request DTOs:**
- `CreateTenantRequest` — full_name (required), email, phone_number, national_id, address (nested AddressDTO), notes
- `UpdateTenantRequest` — all fields optional (partial update)

**Response DTOs:**
- `TenantResponse` — all fields including nested address, landlord_id, timestamps
- `TenantListResponse` — array of TenantResponse + total count for pagination

## 6. HTTP Handler & Routes

### `delivery/http/handler/tenant_handler.go`

Standard CRUD methods + Deactivate, all with Swagger annotations. Extracts user ID from auth context for ownership checks. Follows same patterns as `beneficiary_handler.go`.

### Routes (added to `router.go`)

```
POST   /api/v1/tenants                 — auth + landlord/admin role
GET    /api/v1/tenants                 — auth + landlord/admin role
GET    /api/v1/tenants/{id}            — auth + landlord/admin role
PUT    /api/v1/tenants/{id}            — auth + landlord/admin role
PUT    /api/v1/tenants/{id}/deactivate — auth + landlord/admin role
DELETE /api/v1/tenants/{id}            — auth + landlord/admin role
```

New `TenantHandler` field added to router's `Handlers` struct. Wired in `cmd/server/main.go`.

## 7. Persistence Implementation

### `infrastructure/persistence/pg/tenant_repo_pg.go`

Implements `TenantRepository` interface against PostgreSQL:
- Address fields mapped to/from individual columns
- `List` builds dynamic WHERE clauses from `TenantFilter` (landlord_id, is_active, search via ILIKE on full_name/email/phone)
- `Delete` sets `deleted_at = NOW()`; all queries exclude `deleted_at IS NOT NULL`
- Returns `apperror.NotFound` when tenant not found

## 8. Tests

### Mock: `tests/mocks/tenant_repository_mock.go`

Manual mock following existing pattern (no codegen), with configurable return values per method.

### Unit Tests

| File | What It Tests | Estimated Count |
|:---|:---|:---|
| `tenant_entity_test.go` | Validate(), NewTenant(), contact method rule, email format | ~8 tests |
| `tenant_usecase_test.go` | All 6 use cases: happy paths, ownership checks, email uniqueness, not-found | ~18 tests |
| `tenant_handler_test.go` | HTTP handlers: valid/invalid requests, auth context, response shapes | ~12 tests |

**~38 new tests**, bringing the project total from 88 to ~126.

## Files to Create/Modify

### New Files (10)
1. `backend/internal/domain/entity/address.go`
2. `backend/internal/domain/entity/tenant.go`
3. `backend/internal/domain/repository/tenant_repository.go`
4. `backend/internal/usecase/tenant/create_tenant.go`
5. `backend/internal/usecase/tenant/get_tenant.go`
6. `backend/internal/usecase/tenant/list_tenants.go`
7. `backend/internal/usecase/tenant/update_tenant.go`
8. `backend/internal/usecase/tenant/deactivate_tenant.go`
9. `backend/internal/usecase/tenant/delete_tenant.go`
10. `backend/internal/delivery/http/dto/tenant_dto.go`
11. `backend/internal/delivery/http/handler/tenant_handler.go`
12. `backend/internal/infrastructure/persistence/pg/tenant_repo_pg.go`
13. `backend/migrations/000007_create_tenants.up.sql`
14. `backend/migrations/000007_create_tenants.down.sql`
15. `backend/tests/mocks/tenant_repository_mock.go`
16. `backend/tests/unit/tenant_entity_test.go`
17. `backend/tests/unit/tenant_usecase_test.go`
18. `backend/tests/unit/tenant_handler_test.go`

### Modified Files (3)
1. `backend/internal/domain/entity/errors.go` — add tenant domain errors
2. `backend/internal/delivery/http/router/router.go` — add TenantHandler to Handlers struct + routes
3. `backend/cmd/server/main.go` — wire tenant repo -> use cases -> handler -> Handlers

## Deferred (Not in This Iteration)
- User account auto-creation when creating a tenant (section 5.7 of business logic ref)
- Tenant reactivation (can be added as a follow-up)
- Integration tests (will be added when testcontainers are set up)
