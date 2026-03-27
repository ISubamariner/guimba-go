# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Guimba-GO is a municipal social protection management system (rewrite of a Python/FastAPI app). Go backend + Next.js frontend, using Clean Architecture with polyglot persistence (PostgreSQL, MongoDB, Redis).

**Module**: `github.com/ISubamariner/guimba-go/backend`

## Build & Run Commands

```bash
# Start infrastructure (PostgreSQL, MongoDB, Redis)
docker compose up -d postgres mongodb redis

# Run backend locally (from repo root, reads .env)
cd backend && go run cmd/server/main.go

# Run full stack via Docker
docker compose up -d

# Run all backend tests
cd backend && go test ./tests/...

# Run only unit tests
cd backend && go test ./tests/unit/...

# Run a single test file
cd backend && go test ./tests/unit/ -run TestFunctionName

# Run integration tests (requires DB)
cd backend && go test -tags=integration ./tests/integration/...

# Generate Swagger docs
cd backend && swag init -g cmd/server/main.go -o docs/

# Run database migrations manually
# Migrations auto-run on server startup via golang-migrate

# Frontend
cd frontend && npm install && npm run dev

# Playwright E2E
cd tests/playwright && npx playwright test
```

## Architecture

### Clean Architecture Layers (dependencies point inward only)

```
delivery/http/handler → usecase → domain/repository (interface)
                                        ↑
infrastructure/persistence ─────────────┘ (implements)
```

| Layer | Path | Rules |
|:---|:---|:---|
| **Domain** | `backend/internal/domain/` | Zero external deps (exception: `google/uuid`). Entities, repository interfaces, value objects. |
| **Use Cases** | `backend/internal/usecase/` | One file per use case, grouped by aggregate. Depends only on domain interfaces. |
| **Infrastructure** | `backend/internal/infrastructure/` | Implements domain interfaces. DB drivers, cache, config. PG repos in `persistence/pg/`, Mongo in `persistence/mongo/`. |
| **Delivery** | `backend/internal/delivery/http/` | Handlers, middleware, router, DTOs. Never import infrastructure directly (injected in `main.go`). |

### Dependency Injection

All wiring happens in `backend/cmd/server/main.go`: config → DB pools → repos → use cases → handlers → router. Follow the existing pattern when adding new modules.

### Key Directories

- `backend/pkg/` — Shared utilities (`apperror/`, `logger/`, `auth/`, `validator/`)
- `backend/migrations/` — SQL migration files (`{number}_{description}.up.sql` / `.down.sql`)
- `backend/docs/` — Auto-generated Swagger (do not edit manually)
- `backend/tests/` — All Go tests (unit, integration, e2e) — must be here due to `internal` package visibility
- `backend/tests/mocks/` — Manual mock implementations of repository interfaces

### Current Modules (Phase 4)

Programs, Users & Auth (JWT + RBAC), Beneficiaries (with program enrollment). See `MASTERPLAN.md` Phase 4 for next modules.

## Conventions

- **Go files**: `snake_case.go` | **TypeScript files**: `kebab-case.tsx`
- **DB tables**: `snake_case`, plural | **DB columns**: `snake_case`
- **API routes**: `/api/v1/kebab-case`
- **Commits**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`)
- **Branches**: `master` (prod), `develop` (integration), `feat/<name>`, `fix/<name>`
- **Error responses**: `{ "error": { "code": "...", "message": "...", "details": [] } }` via `pkg/apperror/`
- **Go imports**: group as stdlib, third-party, internal (blank line separated)
- **DTOs**: Always use `delivery/http/dto/` types for HTTP — never expose domain entities directly

## Infrastructure Defaults

| Service | Host | Port | Default Credentials |
|:---|:---|:---|:---|
| PostgreSQL | localhost | 5432 | guimba / guimba_secret / guimba_db |
| MongoDB | localhost | 27017 | guimba / guimba_secret / guimba_db |
| Redis | localhost | 6380 (external) | guimba_secret |
| Backend | localhost | 8080 | — |
| Frontend | localhost | 3000 | — |

Config loaded via Viper from `.env` file + env vars. See `backend/internal/infrastructure/config/config.go`.

## Testing Patterns

- Unit tests use manual mocks from `backend/tests/mocks/` (no codegen framework)
- Table-driven tests preferred
- Test helpers in `backend/tests/helpers/` (assertions, test DB setup)
- Handler tests use `httptest.NewRecorder` + direct handler calls

## Adding a New Domain Module

1. `domain/entity/<name>.go` — struct + validation methods
2. `domain/repository/<name>_repository.go` — interface
3. `usecase/<name>/` — one file per use case
4. `infrastructure/persistence/pg/<name>_repo_pg.go` — PG implementation
5. `delivery/http/dto/<name>_dto.go` — request/response types
6. `delivery/http/handler/<name>_handler.go` — HTTP handler
7. Register routes in `delivery/http/router/router.go`
8. Wire in `cmd/server/main.go`
9. Add Swagger annotations to handler
10. Write tests in `tests/unit/` + `tests/mocks/`

## Guardrails

### Anti-Redundancy
Before creating new code, check if a similar function/handler/component already exists. Extend existing utilities in `pkg/` or `src/lib/` rather than duplicating.

### Bug Tracking
Before debugging any error, search `.github/skills/bug-tracker/references/bug-log.md` for related keywords. After resolving a bug, add an entry with: Issue, Root Cause, Resolution, Files Changed, Prevention.

### Documentation Sync
After completing features, fixes, or refactors, check if docs in `documentation/` or `.github/` need updating. Log changes in `.github/skills/doc-sync/references/changelog.md`.

### SQL Safety
Always use parameterized queries (`$1`, `$2`). Never modify existing migration files — create new ones. Every `.up.sql` needs a reversible `.down.sql`.

## MCP Servers

Configured in `.mcp.json` at project root (added via `claude mcp add --scope project`):
- **postgres** — Query schemas and verify data before writing repository code
- **mongodb** — Read-only inspection of audit logs and document schemas
- **redis** — Inspect cache state and token blocklist
- **context7** — Look up current library API docs instead of guessing
- **playwright** — Browser automation and E2E testing

Database servers require Docker running (`docker compose up -d`).

## Reference Documents

- `MASTERPLAN.md` — Full project plan, phases, and business logic reference
- `documentation/prompts/business-logic-reference.md` — Original system's business rules (source of truth for behavioral parity)
- `documentation/architecture/clean-architecture.md` — Layer rules and DI pattern
- `.github/copilot-instructions.md` — Original Copilot coding standards (preserved for reference)
- `.github/agents/` — Copilot agent workflows (api-builder, db-migrator, frontend-builder, feature-orchestrator) — useful as step-by-step references
- `.github/skills/` — 15 skill reference docs covering testing patterns, auth, error handling, Docker, Swagger, caching, etc.
