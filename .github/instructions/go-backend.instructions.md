---
applyTo: "backend/**/*.go"
---

# Go Backend Instructions

## Clean Architecture Layers
```
delivery/http/handler → usecase → domain/repository (interface)
                                        ↑
infrastructure/persistence ─────────────┘ (implements)
```

**Dependency rule**: Dependencies point INWARD only. Never import an outer layer from an inner layer.

## Domain Layer (`internal/domain/`)
- Contains entities, repository interfaces, service interfaces, value objects
- ZERO external dependencies (no frameworks, no DB drivers)
- Entities are pure Go structs with business validation methods
- Repository interfaces define the contract; implementations live in `infrastructure/`

## Use Case Layer (`internal/usecase/`)
- One file per use case (e.g., `create_program.go`, `get_program.go`)
- Grouped by domain aggregate (`usecase/program/`, `usecase/user/`)
- Depends only on domain interfaces — never on infrastructure directly
- Accepts domain entities, returns domain entities or errors

## Infrastructure Layer (`internal/infrastructure/`)
- Implements domain interfaces (e.g., `persistence/pg/program_repo_pg.go` implements `domain/repository/ProgramRepository`)
- Contains DB connections, Redis client, external API clients, config
- Only layer allowed to import third-party drivers (pgx, mongo-go-driver, go-redis, etc.)
- **PostgreSQL** implementations go in `persistence/pg/`
- **MongoDB** implementations go in `persistence/mongo/`
- Domain interfaces don't know which DB backs them — that's decided here

## Delivery Layer (`internal/delivery/http/`)
- HTTP handlers call use cases, never repositories directly
- DTOs in `dto/` — map between HTTP request/response and domain entities
- Middleware in `middleware/` — auth, CORS, logging, rate limiting
- Routes in `router/` — Chi route definitions

## Handler Rules
- Parse request into DTO, validate, map to domain entity, call use case, map result to response DTO
- Use `chi.URLParam()` for path params
- Always return structured error responses (see AGENTS.md)

## Testing
- Go tests live in `backend/tests/` (not root `tests/`) — required by Go's `internal` package visibility rules
- Unit tests in `backend/tests/unit/` — mock all dependencies
- Integration tests in `backend/tests/integration/` — real DB via testcontainers
- Mocks in `backend/tests/mocks/`
- Playwright E2E tests remain in root `tests/playwright/`

## Imports
- Group imports: stdlib, third-party, internal (separated by blank lines)
- Never use dot imports
