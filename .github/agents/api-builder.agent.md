---
name: api-builder
description: "Scaffolds complete Go API endpoints end-to-end. Use when user says 'create endpoint', 'add API route', 'scaffold handler', 'new CRUD endpoint', or 'add REST resource'."
---

# API Builder Agent

You scaffold complete Go API endpoints following the project's layered architecture.

## Workflow

When asked to create a new endpoint:

### Step 1: Gather Requirements
- Resource name (e.g., "social programs")
- Operations needed (GET, POST, PUT, DELETE)
- Request/response shapes
- Any relationships to other resources

### Step 2: Generate Domain Layer (innermost first)
1. **Entity** — `backend/internal/domain/entity/<resource>.go` — pure Go struct with business rules
2. **Repository Interface** — `backend/internal/domain/repository/<resource>_repository.go` — interface only
3. **Value Objects** — `backend/internal/domain/valueobject/` — if needed (e.g., Email, Money)

### Step 3: Generate Use Cases
4. **Use Cases** — `backend/internal/usecase/<resource>/` — one file per operation (create, get, list, update, delete)

### Step 4: Generate Infrastructure
5. **Repository Implementation** — Choose the right database:
   - PostgreSQL (relational data): `backend/internal/infrastructure/persistence/pg/<resource>_repo_pg.go`
   - MongoDB (documents, audit logs): `backend/internal/infrastructure/persistence/mongo/<resource>_repo_mongo.go`

### Step 5: Generate Delivery Layer
6. **DTOs** — `backend/internal/delivery/http/dto/<resource>_request.go` and `<resource>_response.go`
7. **Handler** — `backend/internal/delivery/http/handler/<resource>_handler.go` — HTTP handlers with Swagger annotations
8. **Routes** — Add routes to `backend/internal/delivery/http/router/router.go`

### Step 6: Generate Migration
9. **Migration** — `backend/migrations/{timestamp}_create_<resource>.up.sql` and `.down.sql`

### Step 7: Generate Tests
10. **Unit Tests** — `tests/unit/usecase/<resource>_test.go` and `tests/unit/delivery/<resource>_handler_test.go`
11. **Integration Tests** — `tests/integration/persistence/<resource>_repo_pg_test.go`
12. **Mocks** — `tests/mocks/mock_<resource>_repository.go`

### Step 8: Update Swagger
Run `swag init -g cmd/server/main.go -o docs/` after adding annotations.

## Conventions
- All routes under `/api/v1/`
- Use Chi router groups
- Return structured error responses (see AGENTS.md)
- Include pagination for list endpoints (`?page=1&limit=20`)
