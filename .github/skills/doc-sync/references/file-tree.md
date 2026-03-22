# Guimba-GO — Project File Tree & File Registry

> **Purpose**: Running inventory of every file in the project with a description of what it does. Updated whenever files are added, renamed, or removed. Use this to quickly orient in the codebase or detect orphaned/missing files.
>
> **Maintained by**: `doc-sync` skill  
> **Last Updated**: 2026-03-22

---

## Legend

| Icon | Meaning |
|:---|:---|
| 📁 | Directory |
| 📄 | File with content |
| 🔒 | Gitkeep placeholder (empty, awaiting implementation) |
| ⚙️ | Configuration file |
| 🐳 | Docker-related |
| 🤖 | Copilot AI configuration |

---

## Root

```
Guimba-GO/
├── .env.example                         ⚙️ Environment variable template
├── .gitignore                           ⚙️ Git exclusion rules
├── AGENTS.md                            📄 Project overview, naming conventions, error format, branch strategy
├── MASTERPLAN.md                        📄 Full implementation roadmap (phases 0–10) with task checklists
├── docker-compose.yml                   🐳 Orchestrates all services: Postgres, MongoDB, Redis, backend, frontend
```

| File | What It Does |
|:---|:---|
| `.env.example` | Template with defaults for `POSTGRES_*`, `MONGO_*`, `REDIS_*`, `JWT_SECRET`, `LOG_LEVEL` |
| `.gitignore` | Excludes binaries, caches, `.env`, IDE configs, OS files |
| `AGENTS.md` | Primary project context loaded by Copilot — architecture diagram, naming rules, error shape, commit/branch conventions |
| `MASTERPLAN.md` | 35KB roadmap: Phase 0 (agentic brain) through Phase 10 (production), plus business logic reference section |
| `docker-compose.yml` | PostgreSQL 16, MongoDB 7, Redis 7, Go backend (port 8080), Next.js frontend (port 3000) with healthchecks and volumes |

---

## .github/ — Copilot AI Configuration

```
.github/
├── copilot-instructions.md              🤖 Global instructions (always loaded)
├── copilot/
│   └── mcp.json                         🤖 MCP server configs (Postgres, Mongo, Redis, Docker, Playwright)
├── instructions/                        🤖 Path-specific instructions
│   ├── database.instructions.md         Applied to: **/*.sql, backend/migrations/**, backend/**/persistence/**
│   ├── go-backend.instructions.md       Applied to: backend/**/*.go
│   └── nextjs-frontend.instructions.md  Applied to: frontend/**/*.{ts,tsx}
├── agents/                              🤖 Custom agents (invoked on demand)
│   ├── api-builder.agent.md             Scaffolds complete Go API endpoints end-to-end
│   ├── db-migrator.agent.md             Handles database schema changes and migrations
│   ├── feature-orchestrator.agent.md    Orchestrates full vertical feature slices (backend + DB + frontend)
│   └── frontend-builder.agent.md        Creates Next.js pages, components, and hooks
└── skills/                              🤖 Auto-detected skills (15 total)
    ├── api-client/
    │   ├── SKILL.md                     Frontend API client management (src/lib/api.ts)
    │   └── references/client-patterns.md    Patterns for typed API calls, auth headers, error handling
    ├── auth-rbac/
    │   ├── SKILL.md                     JWT auth, RBAC, login/register, middleware guards, token refresh
    │   └── references/auth-patterns.md      Auth implementation patterns and middleware examples
    ├── bug-tracker/
    │   ├── SKILL.md                     Bug tracking: "search before debug" rule, Issue→Cause→Fix format
    │   └── references/bug-log.md            Persistent bug history (append-only memory)
    ├── ci-cd/
    │   ├── SKILL.md                     GitHub Actions CI/CD: lint, test, build, deploy pipelines
    │   └── references/workflow-templates.md  Reusable workflow YAML templates
    ├── design-system/
    │   ├── SKILL.md                     CSS tokens, UI primitives, Tailwind theme, visual consistency
    │   └── references/token-registry.md     Complete design token definitions (colors, spacing, typography)
    ├── doc-sync/
    │   ├── SKILL.md                     Keeps all docs in sync with code changes; 5-tier doc registry
    │   └── references/
    │       ├── changelog.md                 Audit trail of all documentation updates
    │       └── file-tree.md                 ← THIS FILE (project inventory)
    ├── docker-compose-services/
    │   ├── SKILL.md                     Docker Compose management for local dev services
    │   └── references/compose-patterns.md   Docker Compose patterns and troubleshooting
    ├── env-config/
    │   ├── SKILL.md                     Environment config management, .env files, secret handling
    │   └── references/config-template.md    Config validation patterns and templates
    ├── error-handling/
    │   ├── SKILL.md                     Standardized error handling via pkg/apperror/
    │   └── references/error-codes.md        Error code registry and HTTP status mapping
    ├── go-testing/
    │   ├── SKILL.md                     Go testing patterns: table-driven, mocking, coverage
    │   ├── scripts/run-tests.ps1            PowerShell script to run tests with coverage
    │   └── references/test-patterns.md      Detailed Go testing idioms and examples
    ├── playwright-testing/
    │   ├── SKILL.md                     E2E tests, visual regression, browser automation
    │   └── references/playwright-patterns.md  Page Object Model patterns, screenshot baselines
    ├── redis-caching/
    │   ├── SKILL.md                     Cache-aside patterns, TTL strategy, key naming, invalidation
    │   └── references/cache-patterns.md     Redis caching implementation patterns
    ├── security-hardening/
    │   ├── SKILL.md                     CORS, CSP, rate limiting, sanitization, OWASP Top 10
    │   └── references/security-checklist.md Security audit checklist
    ├── seed-data/
    │   ├── SKILL.md                     Database seed data and test fixtures
    │   └── references/fixture-templates.md  Seed data templates and fixture patterns
    └── swagger-gen/
        ├── SKILL.md                     Swagger/OpenAPI doc generation from Go handler comments
        └── references/annotation-guide.md   Full swaggo annotation reference
```

---

## backend/ — Go Backend (Clean Architecture)

```
backend/
├── Dockerfile                           🐳 Multi-stage build: golang:1.24-alpine → alpine:3.21
├── go.mod                               ⚙️ Go module: github.com/ISubamarinen/guimba-go/backend (Go 1.26.1)
├── go.sum                               ⚙️ Dependency checksums
├── cmd/
│   └── server/
│       └── main.go                      📄 Application entry point — logger init, HTTP server, graceful shutdown
├── docs/
│   └── .gitkeep                         🔒 Swagger output directory (populated by swag init)
├── migrations/
│   └── .gitkeep                         🔒 SQL migration files ({timestamp}_{desc}.up.sql / .down.sql)
├── internal/
│   ├── delivery/http/                   📁 HTTP layer (outermost)
│   │   ├── dto/
│   │   │   └── .gitkeep                 🔒 Request/response DTOs (mapped to/from domain entities)
│   │   ├── handler/
│   │   │   └── .gitkeep                 🔒 HTTP handlers (call use cases, return DTOs)
│   │   ├── middleware/
│   │   │   └── .gitkeep                 🔒 Auth, CORS, logging, rate limiting middleware
│   │   └── router/
│   │       └── .gitkeep                 🔒 Chi route definitions (/api/v1/*)
│   ├── domain/                          📁 Domain layer (innermost — zero external deps)
│   │   ├── entity/
│   │   │   └── .gitkeep                 🔒 Domain entities (User, Tenant, Property, Debt, Transaction, Role, Permission)
│   │   ├── repository/
│   │   │   └── .gitkeep                 🔒 Repository interfaces (contracts for persistence)
│   │   ├── service/
│   │   │   └── .gitkeep                 🔒 Domain service interfaces
│   │   └── valueobject/
│   │       └── .gitkeep                 🔒 Value objects (Money, Address, UserRole)
│   ├── infrastructure/                  📁 Infrastructure layer (implements domain interfaces)
│   │   ├── cache/
│   │   │   └── .gitkeep                 🔒 Redis cache client and patterns
│   │   ├── config/
│   │   │   └── .gitkeep                 🔒 Application configuration (env parsing, validation)
│   │   ├── database/
│   │   │   └── .gitkeep                 🔒 Database connections (PostgreSQL pgx pool, MongoDB client)
│   │   ├── external/
│   │   │   └── .gitkeep                 🔒 External service clients (email, OCR/Gemini)
│   │   └── persistence/
│   │       ├── pg/
│   │       │   └── .gitkeep             🔒 PostgreSQL repository implementations
│   │       └── mongo/
│   │           └── .gitkeep             🔒 MongoDB repository implementations (audit logs, documents)
│   └── usecase/                         📁 Use case layer (application business logic)
│       ├── program/
│       │   └── .gitkeep                 🔒 Program-related use cases
│       └── user/
│           └── .gitkeep                 🔒 User-related use cases
└── pkg/                                 📁 Shared packages (importable by any layer)
    ├── apperror/
    │   ├── apperror.go                  📄 AppError struct, error codes (NotFound, Validation, etc.), factory functions
    │   └── response.go                  📄 WriteError() — serializes AppError to JSON HTTP response
    ├── logger/
    │   └── logger.go                    📄 slog.Logger setup with JSON output and configurable log level
    └── validator/
        └── validator.go                 📄 Wrapper around go-playground/validator with human-readable messages
```

### Key Backend Files

| File | What It Does |
|:---|:---|
| `cmd/server/main.go` | Bootstraps the app: creates logger, HTTP server with `/health` endpoint, listens on `:8080`, handles OS signals for graceful shutdown |
| `pkg/apperror/apperror.go` | Defines `AppError{Code, Message, HTTPStatus, Details}` with factory functions: `NewNotFound()`, `NewValidation()`, `NewUnauthorized()`, `NewForbidden()`, `NewConflict()`, `NewInternal()` |
| `pkg/apperror/response.go` | `WriteError(w, err)` → writes structured JSON: `{"error":{"code":"...","message":"...","details":[...]}}` |
| `pkg/logger/logger.go` | `New()` → parses `LOG_LEVEL` env var, returns `*slog.Logger` with JSON handler |
| `pkg/validator/validator.go` | `Validate(s)` → runs struct tag validation, returns slice of human-readable error strings |
| `go.mod` | Module path, Go version, dependencies (currently only `validator/v10`) |
| `Dockerfile` | Stage 1: compile binary. Stage 2: copy into minimal alpine image, expose 8080 |

---

## frontend/ — Next.js Frontend

```
frontend/
├── Dockerfile                           🐳 Multi-stage build: node → build → production runner (port 3000)
├── package.json                         ⚙️ Next.js 16.2.0, React 19.2.4, Tailwind CSS 4
├── package-lock.json                    ⚙️ Locked dependency tree
├── tsconfig.json                        ⚙️ TypeScript strict mode config
├── next.config.ts                       ⚙️ Next.js configuration
├── postcss.config.mjs                   ⚙️ PostCSS config for Tailwind
├── eslint.config.mjs                    ⚙️ ESLint config (Next.js preset)
├── .gitignore                           ⚙️ Frontend-specific git exclusions
├── README.md                            📄 Next.js boilerplate readme
├── public/                              📁 Static assets
│   ├── file.svg
│   ├── globe.svg
│   ├── next.svg
│   ├── vercel.svg
│   └── window.svg
├── src/
│   ├── app/                             📁 Next.js App Router pages
│   │   ├── layout.tsx                   📄 Root layout — HTML metadata, Geist fonts, global CSS wrapper
│   │   ├── page.tsx                     📄 Home page — starter template (to be replaced)
│   │   ├── globals.css                  📄 Tailwind base + CSS custom properties for theming
│   │   └── favicon.ico                  Static favicon
│   ├── components/
│   │   └── ui/
│   │       └── .gitkeep                 🔒 UI primitives (Button, Card, Input, Badge, etc.)
│   ├── hooks/
│   │   └── .gitkeep                     🔒 Custom React hooks (useAuth, useDebts, etc.)
│   ├── lib/
│   │   └── .gitkeep                     🔒 API client, utilities (cn(), formatMoney(), etc.)
│   ├── styles/
│   │   └── .gitkeep                     🔒 Design tokens, typography, layout, component CSS
│   └── types/
│       └── .gitkeep                     🔒 TypeScript type definitions for API responses
└── __tests__/                           📁 Frontend tests
    ├── components/                      📁 Component unit tests
    │   └── (empty)
    └── lib/                             📁 Utility/hook unit tests
        └── (empty)
```

### Key Frontend Files

| File | What It Does |
|:---|:---|
| `src/app/layout.tsx` | Root layout: sets page title, loads Geist Sans/Mono fonts, applies dark mode CSS vars |
| `src/app/page.tsx` | Landing page (currently Next.js starter — will be replaced with dashboard) |
| `src/app/globals.css` | Imports Tailwind, defines `--background`/`--foreground` CSS custom properties |
| `package.json` | Scripts: `dev`, `build`, `start`, `lint`. Key deps: next, react, tailwindcss |
| `Dockerfile` | 3-stage build: install deps → build Next.js → run production server |

---

## tests/ — Centralized Test Directory (Go Backend)

```
tests/
├── unit/                                📁 Unit tests (mock all dependencies)
│   ├── domain/
│   │   └── .gitkeep                     🔒 Entity and value object tests
│   ├── usecase/
│   │   └── .gitkeep                     🔒 Use case tests with mocked repositories
│   └── delivery/
│       └── .gitkeep                     🔒 Handler tests with mocked services
├── integration/                         📁 Integration tests (real DB via testcontainers)
│   ├── api/
│   │   └── .gitkeep                     🔒 Full HTTP request/response tests
│   └── persistence/
│       └── .gitkeep                     🔒 Repository tests against real Postgres/Mongo
├── e2e/                                 📁 End-to-end tests
│   └── flows/
│       └── .gitkeep                     🔒 Multi-step user flow tests
├── mocks/
│   └── .gitkeep                         🔒 Mock implementations of domain interfaces
├── helpers/
│   └── .gitkeep                         🔒 Shared test utilities and builders
└── fixtures/
    └── .gitkeep                         🔒 Test data fixtures (JSON, SQL)
```

---

## documentation/ — Project Documentation Hub

```
documentation/
├── README.md                            📄 Documentation index with directory map and quick links
├── project/
│   ├── conventions.md                   📄 Naming conventions, error format, commit strategy
│   ├── setup-guide.md                   📄 Dev environment setup: Docker, backend, frontend, migrations
│   └── tech-stack.md                    📄 Full tech stack with rationale for each choice
├── architecture/
│   ├── clean-architecture.md            📄 4-layer architecture: domain → usecase → infra → delivery
│   ├── design-system.md                 📄 UI primitives, CSS tokens, Tailwind theme, dark mode
│   └── testing-strategy.md              📄 Test pyramid: Go unit/integration, Playwright E2E, frontend unit
├── api/
│   └── README.md                        📄 API documentation hub (Swagger lives in backend/docs/)
├── prompts/
│   ├── business-logic-reference.md      📄 Complete business logic from guimba-debt-tracker (35KB reference)
│   ├── connected-trio-original.md       📄 Original "Connected Trio" prompt that inspired the architecture
│   ├── design-decisions.md              📄 Why Copilot-native skills over custom prompt system
│   └── rebrand-and-extend.prompt.md     📄 Guide to rebrand SPMIS → Guimba + create 8 new skills
└── copilot-config/
    └── README.md                        📄 Index of all Copilot configuration layers
```

---

## Implementation Status

| Layer | Status | Notes |
|:---|:---|:---|
| **Project scaffolding** | ✅ Complete | All directories, configs, Docker, Copilot config |
| **pkg/ utilities** | ✅ Complete | apperror, logger, validator |
| **cmd/server/main.go** | ✅ Complete | HTTP server with graceful shutdown, health endpoint |
| **Domain entities** | 🔒 Scaffolded | .gitkeep placeholders — entities not yet implemented |
| **Domain value objects** | 🔒 Scaffolded | .gitkeep placeholders — Money, Address, UserRole not yet implemented |
| **Repository interfaces** | 🔒 Scaffolded | .gitkeep placeholders |
| **Use cases** | 🔒 Scaffolded | .gitkeep placeholders |
| **Persistence (PG)** | 🔒 Scaffolded | .gitkeep placeholders |
| **Persistence (Mongo)** | 🔒 Scaffolded | .gitkeep placeholders |
| **HTTP handlers** | 🔒 Scaffolded | .gitkeep placeholders |
| **HTTP DTOs** | 🔒 Scaffolded | .gitkeep placeholders |
| **Middleware** | 🔒 Scaffolded | .gitkeep placeholders |
| **Router** | 🔒 Scaffolded | .gitkeep placeholders |
| **Migrations** | 🔒 Scaffolded | .gitkeep placeholder |
| **Frontend pages** | 🟡 Starter only | Next.js boilerplate page, no app pages yet |
| **Frontend components** | 🔒 Scaffolded | .gitkeep placeholders |
| **Tests** | 🔒 Scaffolded | All test directories have .gitkeep placeholders |

---

## Update Protocol

When this file needs updating, use the following rules:

1. **File added**: Add an entry under the correct directory section with a description
2. **File removed**: Remove the entry and note it was deleted
3. **File renamed**: Update the path, keep the description
4. **Directory added**: Add a new subsection if significant, or a line if minor
5. **.gitkeep → real file**: Change 🔒 to 📄 and add a description of what was implemented
6. **Implementation status changed**: Update the Status table at the bottom
