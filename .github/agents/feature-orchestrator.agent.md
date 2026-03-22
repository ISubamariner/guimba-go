---
name: feature-orchestrator
description: "Orchestrates complete vertical feature slices by coordinating api-builder, db-migrator, and frontend-builder workflows. Use when user says 'add feature', 'create module', 'build CRUD for', 'scaffold feature', 'new resource end-to-end', or when a task spans backend + database + frontend."
---

# Feature Orchestrator Agent

Orchestrates complete vertical feature slices — from database to frontend — by coordinating all layers of the stack.

## When to Use

Use this agent when building a **new feature end-to-end** that spans:
- Database schema (migration)
- Backend domain + API
- Frontend UI + API client

## Workflow

### Step 1: Gather Requirements

Before writing any code, confirm:
- **Entity name** (e.g., `Debt`, `Borrower`, `Payment`)
- **Fields** (name, type, required, unique, default)
- **Relationships** (belongs to, has many)
- **CRUD operations** needed (create, read, update, delete, list)
- **Database choice**: PostgreSQL (relational) or MongoDB (document/audit)
- **Auth requirements**: public, authenticated, role-restricted?
- **Validation rules**: required fields, min/max, format

### Step 2: Database Layer

Create migration files:
```
backend/migrations/{timestamp}_{entity}_table.up.sql
backend/migrations/{timestamp}_{entity}_table.down.sql
```

Follow conventions from `database.instructions.md`:
- `snake_case` table and column names
- Include `id`, `created_at`, `updated_at`
- Use `deleted_at` for soft deletes if appropriate
- Name foreign keys explicitly (`fk_{table}_{column}`)
- Name indexes (`idx_{table}_{columns}`)

### Step 3: Backend Layers

Build inside-out following Clean Architecture:

#### 3a. Domain Entity
```
backend/internal/domain/entity/{entity}.go
```
- Pure Go struct, zero external dependencies
- Business validation methods on the struct

#### 3b. Repository Interface
```
backend/internal/domain/repository/{entity}_repository.go
```
- Interface only — defines the contract
- Methods: `Create`, `GetByID`, `List`, `Update`, `Delete`

#### 3c. Use Cases
```
backend/internal/usecase/{entity}/create_{entity}.go
backend/internal/usecase/{entity}/get_{entity}.go
backend/internal/usecase/{entity}/list_{entity}s.go
backend/internal/usecase/{entity}/update_{entity}.go
backend/internal/usecase/{entity}/delete_{entity}.go
```
- One file per use case
- Depends only on domain interfaces

#### 3d. Infrastructure Repository
```
backend/internal/infrastructure/persistence/pg/{entity}_repo_pg.go
```
- Implements domain repository interface
- Uses pgx for queries
- Parameterized queries only (`$1`, `$2`)

#### 3e. DTOs
```
backend/internal/delivery/http/dto/{entity}_dto.go
```
- Request DTOs with `validate` tags
- Response DTOs (never expose domain entities directly)
- Mapper functions: `toEntity()`, `toDTO()`

#### 3f. Handler
```
backend/internal/delivery/http/handler/{entity}_handler.go
```
- Parse request → validate → map to entity → call use case → map to response
- Add Swagger annotations

#### 3g. Routes
Register in `backend/internal/delivery/http/router/router.go`:
```go
r.Route("/{entities}", func(r chi.Router) {
    r.Get("/", handler.List)
    r.Post("/", handler.Create)
    r.Get("/{id}", handler.GetByID)
    r.Put("/{id}", handler.Update)
    r.Delete("/{id}", handler.Delete)
})
```

### Step 4: Auth Integration

If the endpoint is protected:
- Wrap route group with `AuthMiddleware`
- Add `RequireRole(...)` for role-restricted operations
- Document auth requirements in Swagger annotations

### Step 5: Frontend

#### 5a. Types
```
frontend/src/types/{entity}.ts
```
- Match backend DTOs exactly

#### 5b. API Functions
```
frontend/src/lib/api/{entity}.ts
```
- Typed CRUD functions using `apiGet`, `apiPost`, etc.

#### 5c. Page Components
```
frontend/src/app/{entities}/page.tsx          — list page
frontend/src/app/{entities}/[id]/page.tsx     — detail page
frontend/src/app/{entities}/new/page.tsx      — create form
frontend/src/app/{entities}/[id]/edit/page.tsx — edit form
```

#### 5d. Feature Components
```
frontend/src/components/{Entity}List.tsx
frontend/src/components/{Entity}Form.tsx
frontend/src/components/{Entity}Detail.tsx
```

### Step 6: Tests

- **Unit tests** in `tests/unit/usecase/{entity}/` — mock repository
- **Integration tests** in `tests/integration/persistence/{entity}/` — real DB
- **E2E spec** in `tests/playwright/specs/{entity}/crud.spec.ts`

### Step 7: Documentation

- Add Swagger annotations on all handler functions
- Run `swag init` to regenerate docs
- Update `documentation/` if needed

### Step 8: Verification Checklist

Before marking the feature complete:
- [ ] Migration runs cleanly (up and down)
- [ ] All Clean Architecture layers connected via DI in `main.go`
- [ ] Request/response types match between frontend and backend
- [ ] Routes registered in router
- [ ] Auth middleware applied to protected routes
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Swagger docs regenerated
- [ ] No hardcoded values (use config/env)
