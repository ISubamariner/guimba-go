# Guimba-GO Masterplan: Go Backend + Next.js Frontend

## Current Environment Snapshot

| Tool         | Status         |
|-------------|----------------|
| Go           | ✅ 1.26.1       |
| Node.js      | ✅ v24.11.0     |
| npm          | ✅ 11.6.1       |
| Git          | ✅ 2.52.0       |
| Docker       | ✅ 28.0.4       |
| PostgreSQL   | ✅ 16-alpine (Docker, port 5432) |
| MongoDB      | ✅ 7 (Docker, port 27017) |
| Redis        | ✅ 7-alpine (Docker, port 6380) |
| Copilot CLI  | ✅ Active       |
| Custom Instructions/Agents/Skills | ✅ Configured (Phase 0 complete) |

---

## Phase 0: Set Up Your "Agentic Brain" (Copilot Customizations)

The most important step for getting AI-assisted development right.
Create a file-tree of instructions that make Copilot understand your project deeply.

### The 4 Layers of Copilot Customization

| Layer | What It Does | Where It Lives | When Loaded | When To Use |
|-------|-------------|----------------|-------------|-------------|
| **Global Instructions** | Always-on project context (tech stack, conventions, coding standards) | `.github/copilot-instructions.md` | **Every prompt** — always in context | Every project — set and forget |
| **Path-Specific Instructions** | Context scoped to specific file types or directories | `.github/instructions/*.instructions.md` | **When working on matching files** | When different parts of codebase need different rules |
| **Custom Agents** | Specialized personas with specific tools, expertise, and prompts | `.github/agents/*.agent.md` | **On demand** — user invokes or Copilot delegates | Repeating workflows (e.g., "scaffold API endpoint", "create migration") |
| **Skills** | Task-specific instructions + scripts + references, loaded only when relevant | `.github/skills/<name>/SKILL.md` | **Auto-detected** — loaded when Copilot judges it relevant based on description | Complex specialized tasks (e.g., "generate Swagger docs", "debug Docker") |

### How Skills Work: Progressive Disclosure (3 Levels)

Skills use a three-level system to minimize token usage while maintaining expertise:

| Level | What | When Loaded | What Goes Here |
|-------|------|-------------|----------------|
| **Level 1: YAML Frontmatter** | Name + description | **Always** — injected into system prompt | Just enough for Copilot to know WHEN to use this skill |
| **Level 2: SKILL.md Body** | Full instructions | **When skill is triggered** — Copilot decides it's relevant | Step-by-step workflow, examples, error handling |
| **Level 3: Linked Files** | scripts/, references/, assets/ | **On demand** — Copilot navigates to them as needed | Detailed docs, executable scripts, templates |

> **Key Insight**: The description field is the most critical part. It determines whether
> Copilot loads your skill. It must include WHAT the skill does AND WHEN to use it
> (specific trigger phrases users would say).

### Skill Folder Structure & Naming Rules

```
skill-name/                    # ← kebab-case ONLY (no spaces, underscores, or capitals)
├── SKILL.md                   # ← REQUIRED, must be exactly "SKILL.md" (case-sensitive)
├── scripts/                   # ← Optional: executable code (Python, Bash, etc.)
│   └── validate.py
├── references/                # ← Optional: detailed docs loaded on demand
│   └── api-patterns.md
└── assets/                    # ← Optional: templates, icons, etc.
    └── template.md
```

**Critical Rules:**
- ✅ Folder name: `kebab-case` (e.g., `swagger-gen`, `go-testing`)
- ❌ No spaces: `Swagger Gen`
- ❌ No underscores: `swagger_gen`
- ❌ No capitals: `SwaggerGen`
- ❌ No `README.md` inside skill folders (all docs go in SKILL.md or references/)
- ❌ No XML angle brackets (`< >`) in YAML frontmatter (security restriction)
- ❌ No "claude" or "anthropic" in skill names (reserved)
- 📏 Keep SKILL.md under 5,000 words — move detailed content to `references/`

### SKILL.md Format & Best Practices

**Minimal required format:**
```yaml
---
name: your-skill-name          # Required: kebab-case, must match folder name
description: What it does. Use when user asks to [specific trigger phrases].
---
```

**Full format with optional fields:**
```yaml
---
name: your-skill-name
description: What it does and when to use it. Include specific trigger phrases.
license: MIT                   # Optional: for open-source skills
metadata:                      # Optional: custom key-value pairs
  author: Your Name
  version: 1.0.0
---
```

**Writing good descriptions (the most important part):**
```yaml
# ✅ GOOD — specific, includes trigger phrases, mentions file types
description: Generates Swagger/OpenAPI documentation from Go handler comments using swaggo/swag. Use when user says "generate swagger", "create API docs", "update OpenAPI spec", or when working with handler/*.go files.

# ✅ GOOD — clear value proposition with triggers
description: Manages Docker Compose services for local development (PostgreSQL, Redis). Use when user says "start services", "docker compose up", "reset database", or "check container status".

# ❌ BAD — too vague, no triggers
description: Helps with API documentation.

# ❌ BAD — too technical, no user trigger phrases
description: Implements OpenAPI 3.0 specification generation pipeline.
```

**Writing effective instructions (the SKILL.md body):**
```markdown
---
name: swagger-gen
description: [as above]
---

# Swagger Documentation Generator

## Instructions

### Step 1: Verify Swag Installation
Check that `swag` CLI is available: `swag --version`
If not installed: `go install github.com/swaggo/swag/cmd/swag@latest`

### Step 2: Add Swagger Comments
Add annotations to handler functions following this pattern:
[specific code examples]

### Step 3: Generate Docs
Run `swag init -g cmd/server/main.go -o docs/`
Expected output: `docs/swagger.json` and `docs/swagger.yaml` created

## Examples

### Example 1: Document a GET endpoint
User says: "Add swagger docs to the health check endpoint"
Actions:
1. Open `internal/handler/health.go`
2. Add `@Summary`, `@Description`, `@Success`, `@Router` annotations
3. Run `swag init`
Result: Endpoint appears in Swagger UI at `/swagger/index.html`

## Troubleshooting

### Error: "swag: command not found"
Cause: swag CLI not in PATH
Solution: Run `go install github.com/swaggo/swag/cmd/swag@latest` and ensure `$GOPATH/bin` is in PATH
```

**Best practices for instructions:**
- ✅ Be specific and actionable ("Run `python scripts/validate.py`" not "Validate the data")
- ✅ Include error handling for common failures
- ✅ Provide concrete examples with expected output
- ✅ Use progressive disclosure — keep core instructions in SKILL.md, move details to `references/`
- ✅ Put critical instructions at the TOP
- ❌ Don't be vague ("Make sure to validate things properly")

### Complete File Tree to Create

```
Guimba-GO/
├── .github/
│   │
│   ├── copilot-instructions.md            ← LAYER 1: Global project context (ALWAYS loaded)
│   │                                         Tech stack, coding standards, project overview
│   │
│   ├── instructions/                      ← LAYER 2: Path-specific instructions
│   │   ├── go-backend.instructions.md     ← applyTo: "backend/**/*.go"
│   │   ├── nextjs-frontend.instructions.md ← applyTo: "frontend/**/*.{ts,tsx}"
│   │   └── database.instructions.md       ← applyTo: "**/*.sql,backend/migrations/**"
│   │
│   ├── agents/                            ← LAYER 3: Custom agents (on-demand personas)
│   │   ├── api-builder.agent.md           ← Scaffolds Go API endpoints end-to-end
│   │   ├── frontend-builder.agent.md      ← Creates Next.js pages and components
│   │   └── db-migrator.agent.md           ← Handles DB schema changes and migrations
│   │
│   └── skills/                            ← LAYER 4: Skills (auto-detected, progressive disclosure)
│       │
│       ├── docker-compose-services/       ← Category: Workflow Automation
│       │   ├── SKILL.md                   ← Manages Docker Compose for local dev
│       │   └── references/
│       │       └── compose-patterns.md    ← Detailed Docker Compose patterns
│       │
│       ├── swagger-gen/                   ← Category: Document & Asset Creation
│       │   ├── SKILL.md                   ← Generates Swagger/OpenAPI docs from Go
│       │   └── references/
│       │       └── annotation-guide.md    ← Full list of swaggo annotations
│       │
│       ├── go-testing/                    ← Category: Workflow Automation
│       │   ├── SKILL.md                   ← Go testing patterns (table-driven, mocks)
│       │   ├── scripts/
│       │   │   └── run-tests.ps1          ← Script to run tests with coverage
│       │   └── references/
│       │       └── test-patterns.md       ← Detailed Go testing idioms
│       │
│       ├── bug-tracker/                   ← Category: Long-Term Memory (from Connected Trio: Learner/Debugger)
│       │   ├── SKILL.md                   ← Bug tracking process, "search before debug" rule
│       │   └── references/
│       │       └── bug-log.md             ← Persistent bug history (Issue→Cause→Fix)
│       │
│       └── doc-sync/                      ← Category: Long-Term Memory (from Connected Trio: Skill Monitor)
│           ├── SKILL.md                   ← Keeps all docs in sync with code changes
│           └── references/
│               └── changelog.md           ← Audit trail of all documentation updates
│
├── AGENTS.md                              ← High-level project description + conventions
│                                             (treated as primary instructions alongside copilot-instructions.md)
│
└── ... (project code)
```

### Skill Validation Checklist (from Anthropic's Guide)

Use this before considering a skill "done":

**Before you start:**
- [ ] Identified 2-3 concrete use cases for this skill
- [ ] Tools identified (built-in or scripts)
- [ ] Planned folder structure

**During development:**
- [ ] Folder named in kebab-case
- [ ] `SKILL.md` file exists (exact spelling, case-sensitive)
- [ ] YAML frontmatter has `---` delimiters
- [ ] `name` field: kebab-case, no spaces, no capitals
- [ ] `description` includes WHAT it does and WHEN to use it (trigger phrases)
- [ ] No XML tags (`< >`) in frontmatter
- [ ] Instructions are clear and actionable (not vague)
- [ ] Error handling / troubleshooting included
- [ ] Examples provided
- [ ] References clearly linked (not inlined if lengthy)

**After adding to project:**
- [ ] Test: ask Copilot "When would you use the [skill-name] skill?" — it should quote your description
- [ ] Test: does it trigger on obvious tasks? (e.g., "generate swagger docs")
- [ ] Test: does it trigger on paraphrased requests? (e.g., "create API documentation")
- [ ] Test: does it NOT trigger on unrelated topics?
- [ ] Monitor for over/under-triggering and iterate on description

### Tasks

- [x] Initialize git repository in Guimba-GO
- [x] Create `.github/copilot-instructions.md` with project context, anti-redundancy guardrails, bug logging convention, doc-sync rule
- [x] Create `AGENTS.md` at repo root with coding conventions
- [x] Create path-specific instruction files (with proper `applyTo` frontmatter) for backend, frontend, database
- [x] Create custom agents for API building, frontend building, DB migration
- [x] Create skills following best practices:
  - [x] `docker-compose-services/SKILL.md` — with description triggers, steps, troubleshooting
  - [x] `swagger-gen/SKILL.md` — with annotation examples, generation steps
  - [x] `go-testing/SKILL.md` — with table-driven test patterns, coverage scripts, "check bug-log before debugging" rule
  - [x] `bug-tracker/SKILL.md` — persistent bug memory with Issue→Cause→Fix format, escalation rules (from Connected Trio: Learner/Debugger)
  - [x] `doc-sync/SKILL.md` — long-term memory manager, 5-tier doc registry, changelog audit trail (from Connected Trio: Skill Monitor)
- [x] Validate each skill against the checklist above

---

## Business Logic Reference (Original Project)

All business processes, domain rules, entity behaviors, and validation workflows from the original `guimba-debt-tracker` (Python/FastAPI v3.1.0) have been extracted into a comprehensive reference document:

📖 **[`documentation/prompts/business-logic-reference.md`](documentation/prompts/business-logic-reference.md)**

This document is the **single source of truth** for behavioral parity during the Go rewrite. It covers:
- All 8 domain entities with field definitions and methods
- 3 value objects (Money, Address, UserRole)
- 10+ business rule sets (debt creation, payments, refunds, overdue detection, etc.)
- 10 service-layer workflows with step-by-step validation sequences
- Complete auth flow (registration, login, token rotation, password reset, blocklist)
- Background jobs (daily overdue notifications via Celery/Redis)
- Dashboard statistics and audit system specifications
- OCR receipt scanning workflow
- Full API endpoint map (62+ endpoints across 11 modules)
- Error taxonomy mapping

> **Rule**: Before implementing any feature, consult this reference to ensure the Go implementation matches the original business logic.

---

## Phase 1: Install Missing Tools

- [x] Install Go (latest stable, currently 1.24.x) — https://go.dev/dl/
- [x] Install PostgreSQL via Docker (v16+)
- [x] Install MongoDB via Docker (v7+)
- [x] Install Redis via Docker
- [x] Install Postman — https://www.postman.com/downloads/ (or use `swaggo/swag` for Swagger-in-code)
- [x] Verify all installations work

### Recommendation: Use Docker for All Services

Since Docker is already installed, run **all services** as containers — including the Go backend and Next.js frontend, not just the databases.
This avoids cluttering the system and makes the entire stack reproducible via `docker-compose.yml`.

| Service | Docker | Notes |
|:---|:---|:---|
| **Go Backend** | ✅ Dockerfile + Compose service | Multi-stage build (build → scratch/alpine) |
| **Next.js Frontend** | ✅ Dockerfile + Compose service | Multi-stage build (deps → build → runner) |
| **PostgreSQL** | ✅ Compose service | `postgres:16-alpine` |
| **MongoDB** | ✅ Compose service | `mongo:7` |
| **Redis** | ✅ Compose service | `redis:7-alpine` |

### Database Strategy: Polyglot Persistence

Each module chooses the best database for its needs. Domain repository interfaces remain the same regardless of which database backs them — this is the power of Clean Architecture.

| Use Case | Database | Why |
|:---|:---|:---|
| **Relational data** (users, programs, beneficiaries, roles) | PostgreSQL | Strong schema, ACID transactions, foreign keys, complex joins |
| **Audit logs & activity tracking** | MongoDB | Append-heavy, flexible schema, no migrations needed |
| **Document storage** (attachments, unstructured data) | MongoDB | Schema-less, binary data support, flexible documents |
| **CQRS read models** (denormalized views) | MongoDB | Fast reads, no joins needed, schema flexibility |
| **Caching** | Redis | In-memory, sub-ms latency |

**Key Rule**: The `domain/repository/` interfaces don't know which database implements them. A `ProgramRepository` interface is the same whether it's backed by Postgres or MongoDB — the implementation in `infrastructure/persistence/` decides.

---

## Phase 2: Project Structure & Scaffolding

### Architecture: Clean Architecture

The backend follows **Clean Architecture** (Uncle Bob) — dependencies point inward, and each layer only knows about the layer directly inside it.

```
┌──────────────────────────────────────────────┐
│          Frameworks & Drivers (outer)         │
│   HTTP handlers, DB drivers, Redis, CLI       │
├──────────────────────────────────────────────┤
│          Interface Adapters                   │
│   Controllers, Gateways, Presenters           │
├──────────────────────────────────────────────┤
│          Use Cases (Application Logic)        │
│   Application-specific business rules         │
├──────────────────────────────────────────────┤
│          Entities (Domain Core)               │
│   Enterprise-wide business rules & types      │
└──────────────────────────────────────────────┘
        Dependencies point INWARD only →
```

**Key Rules:**
- **Domain** (`domain/`) has ZERO external dependencies — no frameworks, no DB drivers, no HTTP
- **Use Cases** (`usecase/`) depend only on Domain interfaces
- **Infrastructure** (`infrastructure/`) implements Domain interfaces (DB, cache, external APIs)
- **Delivery** (`delivery/`) is the outermost layer (HTTP handlers, middleware, routes)
- Dependencies are injected via **interfaces defined in the domain layer**

### Target Folder Structure

```
Guimba-GO/
├── docker-compose.yml       ← Full stack: Go backend + Next.js frontend + PostgreSQL + MongoDB + Redis
├── .env.example             ← Environment variable template
├── .gitignore
├── AGENTS.md
├── .github/                 ← Copilot customizations (from Phase 0)
│
├── backend/                 ← Go API server
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go               ← Entry point — wires all dependencies
│   │
│   ├── internal/
│   │   ├── domain/                    ← LAYER 1: Entities (innermost, zero dependencies)
│   │   │   ├── entity/               ← Core business structs (Program, User, Beneficiary)
│   │   │   │   ├── program.go
│   │   │   │   └── user.go
│   │   │   ├── repository/           ← Repository INTERFACES (not implementations)
│   │   │   │   ├── program_repository.go
│   │   │   │   └── user_repository.go
│   │   │   ├── service/              ← Service INTERFACES (for use case contracts)
│   │   │   │   └── program_service.go
│   │   │   └── valueobject/          ← Value objects (Email, Money, DateRange, etc.)
│   │   │       └── email.go
│   │   │
│   │   ├── usecase/                   ← LAYER 2: Application business rules
│   │   │   ├── program/              ← Grouped by domain aggregate
│   │   │   │   ├── create_program.go
│   │   │   │   ├── get_program.go
│   │   │   │   └── list_programs.go
│   │   │   └── user/
│   │   │       ├── authenticate.go
│   │   │       └── register.go
│   │   │
│   │   ├── infrastructure/            ← LAYER 3: Frameworks & drivers (implements domain interfaces)
│   │   │   ├── config/               ← App configuration (env vars, viper)
│   │   │   ├── database/             ← DB connection pools & migration runners
│   │   │   │   ├── postgres.go       ← pgx connection pool
│   │   │   │   └── mongodb.go        ← mongo-driver client
│   │   │   ├── cache/                ← Redis client wrapper
│   │   │   │   └── redis.go
│   │   │   ├── persistence/          ← Repository IMPLEMENTATIONS
│   │   │   │   ├── pg/               ← PostgreSQL implementations
│   │   │   │   │   ├── program_repo_pg.go
│   │   │   │   │   └── user_repo_pg.go
│   │   │   │   └── mongo/            ← MongoDB implementations
│   │   │   │       ├── audit_log_repo_mongo.go
│   │   │   │       └── document_repo_mongo.go
│   │   │   └── external/             ← Third-party API clients (if any)
│   │   │
│   │   └── delivery/                  ← LAYER 4: Interface adapters (outermost)
│   │       └── http/
│   │           ├── handler/           ← HTTP handlers (call use cases)
│   │           │   ├── program_handler.go
│   │           │   └── user_handler.go
│   │           ├── middleware/        ← Auth, logging, CORS, rate limiting
│   │           ├── router/            ← Route definitions (Chi)
│   │           │   └── router.go
│   │           └── dto/               ← Request/Response DTOs (decoupled from domain entities)
│   │               ├── program_request.go
│   │               └── program_response.go
│   │
│   ├── pkg/                           ← Shared utilities (importable by any layer)
│   │   ├── apperror/                  ← Custom error types & codes
│   │   ├── logger/                    ← slog wrapper
│   │   └── validator/                 ← Validation helpers
│   │
│   ├── migrations/                    ← SQL migration files (golang-migrate)
│   └── docs/                          ← Swagger generated docs
│
├── tests/                             ← ALL automated tests (centralized)
│   ├── unit/                          ← Unit tests (no external dependencies)
│   │   ├── domain/                    ← Entity & value object tests
│   │   │   ├── program_test.go
│   │   │   └── email_test.go
│   │   ├── usecase/                   ← Use case tests (mocked repositories)
│   │   │   ├── create_program_test.go
│   │   │   └── authenticate_test.go
│   │   └── delivery/                  ← Handler tests (mocked use cases)
│   │       └── program_handler_test.go
│   │
│   ├── integration/                   ← Integration tests (real DB via testcontainers)
│   │   ├── persistence/              ← Repository implementation tests
│   │   │   └── program_repo_pg_test.go
│   │   └── api/                      ← Full HTTP endpoint tests
│   │       └── program_api_test.go
│   │
│   ├── e2e/                           ← End-to-end tests (Go-based API flow tests)
│   │   └── flows/
│   │       └── program_flow_test.go
│   │
│   ├── playwright/                    ← Playwright E2E + visual regression tests
│   │   ├── package.json              ← Playwright dependencies (isolated from frontend)
│   │   ├── playwright.config.ts      ← Playwright configuration
│   │   ├── specs/                    ← Test specs organized by feature
│   │   │   ├── auth/
│   │   │   │   ├── login.spec.ts
│   │   │   │   └── register.spec.ts
│   │   │   ├── programs/
│   │   │   │   ├── crud.spec.ts
│   │   │   │   └── list.spec.ts
│   │   │   └── api-validation/       ← Full-stack tests (browser + API + DB assertions)
│   │   │       └── program-api.spec.ts
│   │   ├── pages/                    ← Page Object Model classes
│   │   │   ├── login.page.ts
│   │   │   ├── dashboard.page.ts
│   │   │   └── programs.page.ts
│   │   ├── fixtures/                 ← Playwright test fixtures & data factories
│   │   │   ├── auth.fixture.ts
│   │   │   └── test-data.ts
│   │   ├── helpers/                  ← API helpers for DB/state setup in tests
│   │   │   ├── api-client.ts        ← Direct API calls for test setup/teardown
│   │   │   └── db-seed.ts           ← Seed data via API before browser tests
│   │   └── snapshots/               ← Visual regression baseline screenshots
│   │       └── .gitkeep
│   │
│   ├── fixtures/                      ← Shared test data (JSON, SQL seeds)
│   │   ├── programs.json
│   │   └── seed.sql
│   │
│   ├── mocks/                         ← Generated/manual mock implementations
│   │   ├── mock_program_repository.go
│   │   └── mock_program_service.go
│   │
│   └── helpers/                       ← Shared test utilities
│       ├── test_db.go                 ← Test database setup/teardown
│       └── assertions.go             ← Custom assertion helpers
│
├── frontend/                          ← Next.js app
│   ├── package.json
│   ├── next.config.js
│   ├── tailwind.config.ts             ← Tailwind config with design tokens
│   ├── src/
│   │   ├── app/                       ← App Router pages
│   │   │   └── globals.css            ← Global styles + CSS custom properties (design tokens)
│   │   ├── components/                ← React components
│   │   │   └── ui/                    ← Primitive UI components (Button, Input, Card, Modal, etc.)
│   │   │       ├── button.tsx
│   │   │       ├── input.tsx
│   │   │       ├── card.tsx
│   │   │       ├── modal.tsx
│   │   │       ├── table.tsx
│   │   │       ├── badge.tsx
│   │   │       └── index.ts           ← Barrel export for all UI primitives
│   │   ├── styles/                    ← Consolidated design system
│   │   │   ├── tokens.css             ← CSS custom properties (colors, spacing, radii, shadows)
│   │   │   ├── typography.css         ← Font families, sizes, weights, line-heights
│   │   │   ├── layouts.css            ← Reusable layout patterns (page-shell, sidebar, grid)
│   │   │   └── components.css         ← @apply-based component classes (.btn, .card, .input, etc.)
│   │   ├── lib/                       ← API client, utilities
│   │   │   ├── api.ts
│   │   │   └── cn.ts                  ← clsx + tailwind-merge utility
│   │   ├── hooks/                     ← Custom React hooks
│   │   └── types/                     ← TypeScript types
│   ├── __tests__/            ← Frontend tests (colocated with Next.js conventions)
│   │   ├── components/
│   │   └── lib/
│   └── public/
│
└── docs/                    ← Swagger generated docs (auto-generated, don't edit)
    └── api/
│
├── documentation/                     ← Consolidated documentation hub (human-readable reference)
│   ├── README.md                      ← Index with quick links
│   ├── project/                       ← Project-level docs
│   │   ├── setup-guide.md            ← Dev environment setup
│   │   ├── tech-stack.md             ← Full tech stack with rationale
│   │   └── conventions.md            ← Naming, commits, branches
│   ├── architecture/                  ← Architecture & design decisions
│   │   ├── clean-architecture.md     ← Clean Architecture layers explained
│   │   ├── testing-strategy.md       ← Test pyramid, types, locations
│   │   └── design-system.md          ← Consolidated CSS & token architecture
│   ├── api/                           ← API documentation
│   │   └── README.md                 ← API overview (Swagger is auto-generated)
│   ├── prompts/                       ← System prompts & AI instructions
│   │   ├── connected-trio-original.md ← Original "Connected Trio" prompt
│   │   └── design-decisions.md       ← Why we chose this architecture
│   └── copilot-config/               ← Mirror of .github/ Copilot config
│       └── README.md                 ← Index of all 4 layers + 7 skills
```

### Clean Architecture Dependency Rules

```
delivery/handler → usecase → domain/repository (interface)
                                    ↑
infrastructure/persistence ─────────┘  (implements the interface)
```

| Layer | Can Depend On | Cannot Depend On |
|:---|:---|:---|
| `domain/` | Nothing (only stdlib) | `usecase/`, `infrastructure/`, `delivery/`, any framework |
| `usecase/` | `domain/` | `infrastructure/`, `delivery/` |
| `infrastructure/` | `domain/` (implements interfaces) | `usecase/`, `delivery/` |
| `delivery/` | `usecase/`, `domain/` (for DTOs/entities) | `infrastructure/` (injected at startup) |

### Testing Strategy

All automated tests live in the centralized `tests/` folder at project root:

| Test Type | Location | What It Tests | Dependencies |
|:---|:---|:---|:---|
| **Unit** | `tests/unit/` | Domain logic, use cases, handlers in isolation | Mocks only (no DB, no network) |
| **Integration** | `tests/integration/` | Repository implementations, API endpoints | Real DB via testcontainers-go |
| **E2E (Go)** | `tests/e2e/` | Full API flow tests | Running backend + DB |
| **E2E (Playwright)** | `tests/playwright/specs/` | Browser UI flows, full-stack validation, visual regression | Running frontend + backend + DB |
| **Frontend Unit** | `frontend/__tests__/` | React components, hooks, API client | Jest/Vitest |

**Test conventions:**
- Unit tests use mocks from `tests/mocks/`
- Integration tests use fixtures from `tests/fixtures/`
- Test helpers (DB setup, custom assertions) live in `tests/helpers/`
- Run all backend tests: `go test ./tests/...`
- Run only unit: `go test ./tests/unit/...`
- Run only integration: `go test -tags=integration ./tests/integration/...`
- Run Playwright E2E: `cd tests/playwright && npx playwright test`
- Run Playwright visual regression: `cd tests/playwright && npx playwright test --update-snapshots` (to update baselines)
- Run Playwright specific spec: `cd tests/playwright && npx playwright test specs/auth/login.spec.ts`

### Key Go Libraries

| Purpose | Library | Why |
|---------|---------|-----|
| HTTP Router | `chi` or `gin` | Lightweight, idiomatic Go |
| PostgreSQL | `pgx` (jackc/pgx) | Best pure-Go Postgres driver |
| MongoDB | `mongo-go-driver` (mongodb/mongo-go-driver) | Official MongoDB Go driver |
| Redis | `go-redis/redis` | Most popular Go Redis client |
| Migrations | `golang-migrate/migrate` | CLI + library, SQL-file based |
| Swagger | `swaggo/swag` | Generates Swagger from Go comments |
| Config | `viper` or `env` | Environment variable management |
| Auth/JWT | `golang-jwt/jwt` | Standard JWT library |
| Validation | `go-playground/validator` | Struct validation tags |
| E2E Testing | `playwright` (npm) | Browser E2E, full-stack validation, visual regression |

### Tasks

- [x] Create `docker-compose.yml` with PostgreSQL + MongoDB + Redis services
- [x] Initialize Go module (`go mod init`)
- [x] Scaffold backend Clean Architecture directories (`domain/`, `usecase/`, `infrastructure/`, `delivery/`)
- [x] Scaffold `tests/` directory (unit, integration, e2e, playwright, fixtures, mocks, helpers)
- [x] Initialize Next.js app in `frontend/`
- [ ] Initialize Playwright in `tests/playwright/` (`npm init playwright@latest`)
- [x] Create `.env.example` with all required env vars
- [x] Create `pkg/apperror/` with custom error types
- [x] Create Dockerfiles for backend and frontend
- [x] Add backend + frontend services to `docker-compose.yml`

---

## Phase 3: Core Backend Setup

- [ ] Set up `internal/infrastructure/config/` — read env vars with viper
- [ ] Set up `internal/infrastructure/database/` — pgx connection pool + MongoDB client
- [ ] Set up `internal/infrastructure/cache/` — Redis client
- [ ] Set up `internal/infrastructure/persistence/pg/` — PostgreSQL repository implementations
- [ ] Set up `internal/infrastructure/persistence/mongo/` — MongoDB repository implementations
- [ ] Set up `internal/delivery/http/router/` — Chi router with middleware (CORS, logging, auth)
- [ ] Set up `internal/delivery/http/middleware/` — Auth, CORS, request logging
- [ ] Create `pkg/apperror/` — structured error types matching the API error format
- [ ] Create first migration (schema for Guimba entities)
- [ ] Set up Swagger generation
- [ ] Write health-check endpoint as first use case → handler flow
- [ ] Set up `tests/helpers/test_db.go` — testcontainers-based DB for integration tests
- [ ] Write first unit test in `tests/unit/`

---

## Phase 4: Build Domain Modules (Iterative, per aggregate)

Build module-by-module, following Clean Architecture layers:

For **each** domain aggregate (e.g., Debts, Users, Borrowers):
- [ ] Define domain models and schema
- [ ] Create `domain/entity/` structs (zero dependencies)
- [ ] Create `domain/repository/` interface
- [ ] Create `usecase/` implementations (application business rules)
- [ ] Create `infrastructure/persistence/` implementation (pgx queries)
- [ ] Create `delivery/http/dto/` request/response types
- [ ] Create `delivery/http/handler/` HTTP handlers
- [ ] Register routes in `delivery/http/router/`
- [ ] Add Swagger annotations
- [ ] Write unit tests in `tests/unit/`
- [ ] Write integration tests in `tests/integration/`

Port order:
1. [ ] **Programs** (core entity, simplest CRUD)
2. [ ] **Users & Auth** (authentication, JWT, roles)
3. [ ] **Beneficiaries** (relationships to programs)
4. [ ] **Remaining modules** (iterate)

---

## Phase 5: Frontend (Next.js)

- [ ] Set up Next.js with TypeScript, Tailwind CSS
- [ ] **Design System Foundation**:
  - [ ] Create `src/styles/tokens.css` — CSS custom properties (colors, spacing, radii, shadows)
  - [ ] Create `src/styles/typography.css` — font scale, weights, line-heights
  - [ ] Create `src/styles/layouts.css` — reusable layout patterns (page shell, sidebar, grid)
  - [ ] Create `src/styles/components.css` — `@apply`-based component classes (`.btn`, `.card`, `.input`)
  - [ ] Configure `tailwind.config.ts` to extend theme with design tokens
  - [ ] Create `src/lib/cn.ts` — clsx + tailwind-merge utility
- [ ] **UI Primitive Components** (`src/components/ui/`):
  - [ ] Button (variants: primary, secondary, outline, ghost, destructive; sizes: sm, md, lg)
  - [ ] Input, Textarea, Select (consistent form styling)
  - [ ] Card (header, body, footer slots)
  - [ ] Modal / Dialog
  - [ ] Table (sortable, pagination-ready)
  - [ ] Badge, Alert, Toast
  - [ ] Barrel export in `src/components/ui/index.ts`
- [ ] Create API client layer
- [ ] Port UI pages for Guimba (using design system components)
- [ ] Connect to Go backend APIs
- [ ] Set up Playwright Page Object Models for completed pages
- [ ] Write Playwright E2E specs for critical user flows
- [ ] Establish visual regression baselines for key pages

---

## Things To Remember

1. **Environment Variables Management** — Use `.env` files + `viper` or `godotenv`
2. **CORS Configuration** — Go backend needs CORS middleware for Next.js frontend
3. **Database Migrations Strategy** — Use `golang-migrate` with SQL files, not ORM auto-migrations
4. **API Versioning** — Use `/api/v1/...` prefix from the start
5. **Error Handling Pattern** — Go has no exceptions; design an error response struct early
6. **Logging** — Use `slog` (stdlib in Go 1.21+) or `zerolog`
7. **Testing Strategy** — Centralized `tests/` folder; table-driven unit tests, testcontainers for integration, Playwright for browser E2E + visual regression, all routed through `go test ./tests/...` (Go) or `npx playwright test` (browser)
8. **Hot Reload for Go** — Use `air` (cosmtrek/air) for live-reloading during dev
9. **CI/CD Pipeline** — GitHub Actions workflow for lint + test + build
10. **Rate Limiting** — If this is a public-facing API, add rate limiting middleware
11. **Connection Pooling** — `pgx` handles this natively; Redis has built-in pool in `go-redis`
