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
├── go.mod                               ⚙️ Go module: github.com/ISubamariner/guimba-go/backend (Go 1.26.1)
├── go.sum                               ⚙️ Dependency checksums
├── cmd/
│   └── server/
│       └── main.go                      📄 Entry point — config, DB, cache, router, graceful shutdown
├── docs/
│   ├── docs.go                          📄 Swagger embedded Go file (auto-generated by swag)
│   ├── swagger.json                     📄 OpenAPI spec (JSON, auto-generated)
│   └── swagger.yaml                     📄 OpenAPI spec (YAML, auto-generated)
├── migrations/
│   ├── 000001_create_programs.up.sql    📄 Programs table: UUID PK, soft deletes, auto-updated timestamps
│   ├── 000001_create_programs.down.sql  📄 Reverse: drop programs table, trigger, function
│   ├── 000002_create_roles_permissions.up.sql   📄 Roles, permissions, role_permissions tables
│   ├── 000002_create_roles_permissions.down.sql 📄 Reverse: drop role_permissions, permissions, roles
│   ├── 000003_create_users.up.sql       📄 Users table (UUID, email, hashed_password) + user_roles junction
│   ├── 000003_create_users.down.sql     📄 Reverse: drop user_roles, users
│   ├── 000004_seed_system_roles.up.sql  📄 Seed 3 roles (admin/staff/viewer) + 13 permissions + role_permissions
│   ├── 000004_seed_system_roles.down.sql 📄 Reverse: delete seeded role_permissions, permissions, roles
│   ├── 000005_create_beneficiaries.up.sql 📄 Beneficiaries table + program_beneficiaries junction (many-to-many)
│   └── 000005_create_beneficiaries.down.sql 📄 Reverse: drop program_beneficiaries, beneficiaries
├── internal/
│   ├── delivery/http/                   📁 HTTP layer (outermost)
│   │   ├── dto/
│   │   │   ├── health_response.go       📄 HealthResponse DTO with service status map
│   │   │   ├── program_dto.go           📄 CreateProgramRequest, UpdateProgramRequest, ProgramResponse, ProgramListResponse DTOs with entity mapping
│   │   │   ├── user_dto.go             📄 RegisterRequest, LoginRequest, RefreshRequest, AuthResponse, UserResponse, AssignRoleRequest DTOs
│   │   │   └── beneficiary_dto.go      📄 CreateBeneficiaryRequest, UpdateBeneficiaryRequest, EnrollProgramRequest, BeneficiaryResponse, BeneficiaryListResponse DTOs
│   │   ├── handler/
│   │   │   ├── health_handler.go        📄 Health check handler — pings Postgres, MongoDB, Redis
│   │   │   ├── program_handler.go       📄 Programs CRUD handler — Create, Get, List, Update, Delete with Swagger annotations
│   │   │   ├── auth_handler.go          📄 Auth handler — Register, Login, RefreshToken, GetProfile, Logout with Swagger annotations
│   │   │   ├── user_handler.go          📄 User management handler — List, Update, Delete, AssignRole (admin-only)
│   │   │   └── beneficiary_handler.go   📄 Beneficiary handler — CRUD + EnrollInProgram + RemoveFromProgram with Swagger annotations
│   │   ├── middleware/
│   │   │   ├── auth.go                  📄 AuthMiddleware (JWT validation + blocklist check), RequireRole middleware
│   │   │   ├── logger.go                📄 Request logging middleware (slog, method, path, status, duration)
│   │   │   ├── recovery.go              📄 Panic recovery middleware — returns structured error via apperror
│   │   │   └── request_id.go            📄 X-Request-ID injection (UUID, propagated via context)
│   │   └── router/
│   │       └── router.go                📄 Chi router: middleware stack, CORS, /health, /swagger/*, /api/v1/programs, /auth, /users, /beneficiaries with auth + role guards
│   ├── domain/                          📁 Domain layer (innermost — zero external deps)
│   │   ├── entity/
│   │   │   ├── program.go               📄 Program entity: UUID ID, Name, Description, Status, dates, Validate(), NewProgram()
│   │   │   ├── user.go                  📄 User entity: UUID ID, email, hashed_password, Roles, Validate(), HasRole(), HasPermission()
│   │   │   ├── beneficiary.go            📄 Beneficiary entity: UUID ID, FullName, Email, Phone, NationalID, Address, DOB, Status, ProgramEnrollment
│   │   │   ├── role.go                  📄 Role + Permission entities with validation
│   │   │   └── errors.go                📄 Domain error sentinels: Program, User, Role, Beneficiary errors
│   │   ├── repository/
│   │   │   ├── program_repository.go    📄 ProgramRepository interface: Create, GetByID, List, Update, Delete + ProgramFilter
│   │   │   ├── user_repository.go       📄 UserRepository interface: Create, GetByID, GetByEmail, List, Update, Delete
│   │   │   ├── role_repository.go       📄 RoleRepository interface: GetByID, GetByName, List, GetPermissionsByRoleID
│   │   │   └── beneficiary_repository.go 📄 BeneficiaryRepository interface: CRUD + EnrollInProgram, RemoveFromProgram + BeneficiaryFilter
│   │   ├── service/                     📁 (empty — awaiting domain service interfaces)
│   │   └── valueobject/                 📁 (empty — awaiting value objects)
│   ├── infrastructure/                  📁 Infrastructure layer (implements domain interfaces)
│   │   ├── cache/
│   │   │   ├── redis.go                 📄 NewRedisClient() — go-redis with pooling, ping check
│   │   │   └── token_blocklist.go       📄 TokenBlocklist: Redis-backed JWT blocklist (Block, IsBlocked) with TTL
│   │   ├── config/
│   │   │   └── config.go                📄 Viper-based config: App, Postgres, Mongo, Redis, JWT structs
│   │   ├── database/
│   │   │   ├── postgres.go              📄 NewPostgresPool() — pgx pool (25 max conns, health checks)
│   │   │   ├── mongodb.go               📄 NewMongoClient() — mongo-driver with pooling, ping check
│   │   │   └── migrate.go               📄 RunMigrations() — golang-migrate, file-based SQL migrations
│   │   ├── external/
│   │   │   └── .gitkeep                 🔒 External service clients (email, OCR/Gemini)
│   │   └── persistence/
│   │       ├── pg/
│   │       │   ├── program_repo_pg.go   📄 ProgramRepoPG: PostgreSQL implementation of ProgramRepository (CRUD, soft delete, ILIKE search, pagination)
│   │       │   ├── user_repo_pg.go      📄 UserRepoPG: PostgreSQL implementation of UserRepository (CRUD, eager load roles+permissions)
│   │       │   ├── role_repo_pg.go      📄 RoleRepoPG: PostgreSQL implementation of RoleRepository (role+permission queries)
│   │       │   └── beneficiary_repo_pg.go 📄 BeneficiaryRepoPG: PostgreSQL implementation of BeneficiaryRepository (CRUD, enrollment, ILIKE search, pagination)
│   │       └── mongo/                   📁 (empty — awaiting MongoDB repository implementations)
│   └── usecase/                         📁 Use case layer (application business logic)
│       ├── program/                     📁 Programs use cases (5 files)
│       │   ├── create_program.go        📄 CreateProgramUseCase — validates + persists a new program
│       │   ├── get_program.go           📄 GetProgramUseCase — retrieves by ID, returns NotFound if missing
│       │   ├── list_programs.go         📄 ListProgramsUseCase — paginated list with status/search filter, limit cap (100)
│       │   ├── update_program.go        📄 UpdateProgramUseCase — verifies exists, validates, updates
│       │   └── delete_program.go        📄 DeleteProgramUseCase — verifies exists, soft-deletes
│       ├── auth/                        📁 Auth use cases (4 files)
│       │   ├── register.go              📄 RegisterUseCase — validates, hashes password, creates user, assigns default role
│       │   ├── login.go                 📄 LoginUseCase — verifies credentials, generates JWT token pair
│       │   ├── refresh_token.go         📄 RefreshTokenUseCase — validates refresh token, rotates token pair, blocklists old
│       │   └── get_profile.go           📄 GetProfileUseCase — retrieves user by ID with roles
│       ├── user/                        📁 User management use cases (4 files)
│       │   ├── list_users.go            📄 ListUsersUseCase — paginated user list
│       │   ├── update_user.go           📄 UpdateUserUseCase — update user profile fields
│       │   ├── delete_user.go           📄 DeleteUserUseCase — soft-delete user
│       │   └── assign_role.go           📄 AssignRoleUseCase — assign/remove role from user
│       └── beneficiary/                 📁 Beneficiary use cases (7 files)
│           ├── create_beneficiary.go    📄 CreateBeneficiaryUseCase — validates + persists a new beneficiary
│           ├── get_beneficiary.go       📄 GetBeneficiaryUseCase — retrieves by ID with program enrollments
│           ├── list_beneficiaries.go    📄 ListBeneficiariesUseCase — paginated list with status/program/search filter
│           ├── update_beneficiary.go    📄 UpdateBeneficiaryUseCase — verifies exists, validates, updates
│           ├── delete_beneficiary.go    📄 DeleteBeneficiaryUseCase — verifies exists, soft-deletes
│           ├── enroll_in_program.go     📄 EnrollInProgramUseCase — verifies both exist, creates enrollment
│           └── remove_from_program.go   📄 RemoveFromProgramUseCase — removes enrollment
├── tests/                               📁 Go backend tests (within module for internal/ access)
│   ├── unit/
│   │   ├── apperror_test.go             📄 9 tests: error constructors, HTTP status, WriteError, Unwrap
│   │   ├── middleware_test.go           📄 Tests: RequestID generation/passthrough, Recovery, HealthResponse DTO
│   │   ├── program_entity_test.go       📄 8 tests: NewProgram valid/invalid, name validation, status validation, date range
│   │   ├── program_usecase_test.go      📄 10 tests: CRUD use cases — success, not-found, validation, repo errors, pagination caps
│   │   ├── program_handler_test.go      📄 8 tests: HTTP handlers — create/get/list/delete success, invalid JSON, validation, not-found
│   │   ├── user_auth_test.go            📄 12 tests: register, login, refresh, profile, logout, JWT validation, blocklist
│   │   ├── user_usecase_test.go         📄 10 tests: list users, update, delete, assign/remove role, admin checks
│   │   ├── beneficiary_entity_test.go   📄 9 tests: NewBeneficiary valid/invalid, name validation, status validation, contact required
│   │   ├── beneficiary_usecase_test.go  📄 14 tests: CRUD use cases, enroll/remove program, not-found, pagination
│   │   └── beneficiary_handler_test.go  📄 8 tests: HTTP handlers — create/get/list/delete/enroll success, invalid JSON, validation, not-found
│   ├── mocks/
│   │   ├── program_repository_mock.go   📄 ProgramRepositoryMock: function-based mock for all ProgramRepository methods
│   │   ├── user_repository_mock.go      📄 UserRepositoryMock: function-based mock for all UserRepository methods
│   │   ├── role_repository_mock.go      📄 RoleRepositoryMock: function-based mock for all RoleRepository methods
│   │   ├── token_blocklist_mock.go      📄 TokenBlocklistMock: function-based mock for Block/IsBlocked
│   │   └── beneficiary_repository_mock.go 📄 BeneficiaryRepositoryMock: function-based mock for all BeneficiaryRepository methods
│   └── helpers/
│       ├── test_db.go                   📄 TestDB helper: pgxpool connection, TruncateTable cleanup
│       └── assertions.go               📄 AssertStatus, AssertJSONKey test assertion helpers
└── pkg/                                 📁 Shared packages (importable by any layer)
    ├── apperror/
    │   ├── apperror.go                  📄 AppError struct, error codes (NotFound, Validation, etc.), factory functions
    │   └── response.go                  📄 WriteError() — serializes AppError to JSON HTTP response
    ├── auth/
    │   ├── jwt.go                       📄 JWTManager: GenerateTokenPair (access 15min + refresh 7d), ValidateToken, Claims struct
    │   └── password.go                  📄 HashPassword (bcrypt), CheckPassword — never store plaintext
    ├── logger/
    │   └── logger.go                    📄 slog.Logger setup with JSON output and configurable log level
    └── validator/
        └── validator.go                 📄 Wrapper around go-playground/validator with human-readable messages
```

### Key Backend Files

| File | What It Does |
|:---|:---|
| `cmd/server/main.go` | Bootstraps the app: loads config (Viper), connects to Postgres/MongoDB/Redis, runs migrations, wires program + auth + user + beneficiary modules via DI, starts Chi router with middleware, handles graceful shutdown |
| `internal/infrastructure/config/config.go` | `Load()` → reads `.env` + env vars via Viper, builds typed Config struct, auto-constructs DSN/URI if not provided, validates |
| `internal/infrastructure/database/postgres.go` | `NewPostgresPool()` → pgx pool with 25 max conns, 5 min conns, health check period, ping verification |
| `internal/infrastructure/database/mongodb.go` | `NewMongoClient()` → mongo-driver client with pooling, ping check against primary |
| `internal/infrastructure/database/migrate.go` | `RunMigrations()` → golang-migrate file-based runner, logs version on completion |
| `internal/infrastructure/cache/redis.go` | `NewRedisClient()` → go-redis with 25 pool size, ping verification |
| `internal/delivery/http/router/router.go` | Chi router: RequestID → Recovery → Logger → RealIP → Compress → CORS → routes (/health, /swagger/*, /api/v1/auth, /api/v1/programs, /api/v1/users, /api/v1/beneficiaries) with auth middleware and role guards |
| `internal/delivery/http/handler/health_handler.go` | `Health()` → pings Postgres, MongoDB, Redis; returns `{status, timestamp, services}` |
| `internal/delivery/http/handler/program_handler.go` | Programs CRUD: Create (201), Get (200), List (200, paginated), Update (200), Delete (204) with Swagger annotations |
| `internal/delivery/http/handler/auth_handler.go` | Auth endpoints: Register (201), Login (200), RefreshToken (200), GetProfile (200), Logout (200) |
| `internal/delivery/http/handler/user_handler.go` | User management: List (200), Update (200), Delete (204), AssignRole (200) — admin-only |
| `internal/delivery/http/handler/beneficiary_handler.go` | Beneficiary CRUD + enrollment: Create (201), Get (200), List (200), Update (200), Delete (204), EnrollInProgram (200), RemoveFromProgram (204) |
| `internal/delivery/http/dto/program_dto.go` | CreateProgramRequest, UpdateProgramRequest (with validation tags), ProgramResponse, ProgramListResponse; entity mapping helpers |
| `internal/delivery/http/dto/user_dto.go` | RegisterRequest, LoginRequest, RefreshRequest, AuthResponse, UserResponse, AssignRoleRequest; entity mapping helpers |
| `internal/delivery/http/dto/beneficiary_dto.go` | CreateBeneficiaryRequest, UpdateBeneficiaryRequest, EnrollProgramRequest, BeneficiaryResponse, BeneficiaryListResponse; entity mapping helpers |
| `internal/delivery/http/middleware/auth.go` | AuthMiddleware (JWT + blocklist), RequireRole (role guard) |
| `internal/domain/entity/program.go` | Program struct (UUID, name, description, status, dates, timestamps), NewProgram(), Validate(), ProgramStatus enum |
| `internal/domain/entity/user.go` | User struct (UUID, email, hashed_password, roles), Validate(), HasRole(), HasPermission() |
| `internal/domain/entity/role.go` | Role + Permission entities with validation |
| `internal/domain/entity/errors.go` | Domain error sentinels for Program, User, Role, Beneficiary |
| `internal/domain/repository/program_repository.go` | ProgramRepository interface (Create, GetByID, List, Update, Delete), ProgramFilter struct |
| `internal/domain/repository/user_repository.go` | UserRepository interface (Create, GetByID, GetByEmail, List, Update, Delete) |
| `internal/domain/repository/role_repository.go` | RoleRepository interface (GetByID, GetByName, List, GetPermissionsByRoleID) |
| `internal/usecase/program/*.go` | 5 use cases: CreateProgram, GetProgram, ListPrograms (pagination caps), UpdateProgram, DeleteProgram |
| `internal/usecase/auth/*.go` | 4 use cases: Register, Login, RefreshToken, GetProfile |
| `internal/usecase/user/*.go` | 4 use cases: ListUsers, UpdateUser, DeleteUser, AssignRole |
| `internal/usecase/beneficiary/*.go` | 7 use cases: CreateBeneficiary, GetBeneficiary, ListBeneficiaries, UpdateBeneficiary, DeleteBeneficiary, EnrollInProgram, RemoveFromProgram |
| `internal/infrastructure/persistence/pg/program_repo_pg.go` | ProgramRepoPG: parameterized queries, soft deletes, ILIKE search, pagination |
| `internal/infrastructure/persistence/pg/user_repo_pg.go` | UserRepoPG: eager-loaded roles+permissions, parameterized queries, soft deletes |
| `internal/infrastructure/persistence/pg/role_repo_pg.go` | RoleRepoPG: role/permission queries with junction table JOINs |
| `internal/infrastructure/persistence/pg/beneficiary_repo_pg.go` | BeneficiaryRepoPG: CRUD, enrollment, ILIKE search, pagination, program_beneficiaries JOINs |
| `internal/infrastructure/cache/token_blocklist.go` | Redis SET with TTL matching token remaining lifetime |
| `internal/delivery/http/middleware/` | auth.go, logger.go (request logging), request_id.go (UUID injection), recovery.go (panic → structured error) |
| `migrations/000001_create_programs.up.sql` | Creates programs table with UUID PK, soft deletes, `updated_at` trigger |
| `migrations/000002_create_roles_permissions.*.sql` | Creates roles, permissions, role_permissions tables |
| `migrations/000003_create_users.*.sql` | Creates users table, user_roles junction table |
| `migrations/000004_seed_system_roles.*.sql` | Seeds admin/staff/viewer roles + 13 permissions + role_permissions mappings |
| `migrations/000005_create_beneficiaries.*.sql` | Creates beneficiaries table + program_beneficiaries junction (many-to-many with programs) |
| `pkg/auth/jwt.go` | JWTManager: GenerateTokenPair (access 15min, refresh 7d, HS256), ValidateToken, Claims struct |
| `pkg/auth/password.go` | HashPassword (bcrypt.DefaultCost), CheckPassword — wraps golang.org/x/crypto/bcrypt |
| `pkg/apperror/apperror.go` | `AppError{Code, Message, HTTPStatus, Details}` with 8 factory functions (incl. NewNotFoundMsg) |
| `pkg/apperror/response.go` | `WriteError(w, err)` → structured JSON: `{"error":{...}}` |
| `pkg/logger/logger.go` | `New()` → parses `LOG_LEVEL` env var, returns `*slog.Logger` with JSON handler |
| `pkg/validator/validator.go` | `Validate(s)` → runs struct tag validation, returns slice of human-readable error strings |
| `go.mod` | Module path, Go 1.26.1, dependencies: chi, pgx, mongo-driver, go-redis, viper, jwt, swag, migrate, validator |
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

## tests/ — Cross-Cutting Test Directory

> **Note**: Go backend tests live in `backend/tests/` (required by Go's `internal` package rules). This root `tests/` directory holds Playwright E2E tests, shared fixtures, and cross-cutting resources.

```
tests/
├── unit/                                📁 (legacy scaffolding — Go unit tests moved to backend/tests/unit/)
│   ├── domain/
│   │   └── .gitkeep                     🔒
│   ├── usecase/
│   │   └── .gitkeep                     🔒
│   └── delivery/
│       └── .gitkeep                     🔒
├── integration/                         📁 (legacy scaffolding — Go integration tests will go in backend/tests/integration/)
│   ├── api/
│   │   └── .gitkeep                     🔒
│   └── persistence/
│       └── .gitkeep                     🔒
├── e2e/                                 📁 End-to-end tests
│   └── flows/
│       └── .gitkeep                     🔒 Multi-step user flow tests
├── mocks/
│   └── .gitkeep                         🔒 (Go mocks will go in backend/tests/mocks/)
├── fixtures/
│   └── .gitkeep                         🔒 Shared test data fixtures (JSON, SQL)
└── helpers/                             📁 (empty — Go helpers moved to backend/tests/helpers/)
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
| **Infrastructure: config** | ✅ Complete | Viper-based config loading from env vars/.env |
| **Infrastructure: database** | ✅ Complete | pgx pool (Postgres), mongo-driver (MongoDB), golang-migrate runner |
| **Infrastructure: cache** | ✅ Complete | go-redis client with connection pooling |
| **HTTP middleware** | ✅ Complete | Request logging, request ID, panic recovery, CORS |
| **HTTP router** | ✅ Complete | Chi router with full middleware stack, /health, /swagger/*, /api/v1/programs, /auth, /users, /beneficiaries with auth+role guards |
| **Health endpoint** | ✅ Complete | Pings Postgres, MongoDB, Redis; returns structured response |
| **Swagger docs** | ✅ Complete | swaggo/swag, UI at /swagger/index.html, includes Programs + Auth + Users + Beneficiaries endpoints |
| **Migrations** | ✅ Active | 000001 (programs), 000002 (roles+permissions), 000003 (users), 000004 (seed system roles), 000005 (beneficiaries + program_beneficiaries) |
| **cmd/server/main.go** | ✅ Complete | Config → DB → cache → migrations → program + auth + user + beneficiary module DI → router → graceful shutdown |
| **Go unit tests** | ✅ Active | 88 tests in backend/tests/unit/ (apperror, middleware, health DTO, program, auth, user, beneficiary) |
| **Go test helpers** | ✅ Complete | TestDB, AssertStatus, AssertJSONKey in backend/tests/helpers/ |
| **Go test mocks** | ✅ Active | ProgramRepositoryMock, UserRepositoryMock, RoleRepositoryMock, TokenBlocklistMock, BeneficiaryRepositoryMock |
| **Domain entities** | ✅ Programs + Users + Beneficiaries done | Program, User, Role, Permission, Beneficiary, ProgramEnrollment entities with validation + domain error sentinels |
| **Domain value objects** | 🔒 Not started | Empty directories |
| **Repository interfaces** | ✅ Programs + Users + Beneficiaries done | ProgramRepository, UserRepository, RoleRepository, BeneficiaryRepository |
| **Use cases** | ✅ Programs + Auth + Users + Beneficiaries done | 5 program + 4 auth + 4 user + 7 beneficiary use cases |
| **Persistence (PG)** | ✅ Programs + Users + Beneficiaries done | ProgramRepoPG, UserRepoPG, RoleRepoPG, BeneficiaryRepoPG |
| **Persistence (Mongo)** | 🔒 Not started | Empty directories |
| **HTTP handlers** | ✅ Programs + Auth + Users + Beneficiaries done | Health + Program + Auth + User + Beneficiary handlers with Swagger annotations |
| **HTTP DTOs** | ✅ Programs + Auth + Users + Beneficiaries done | Health, Program, Auth/User, Beneficiary request/response DTOs |
| **Auth middleware** | ✅ Complete | AuthMiddleware (JWT + Redis blocklist), RequireRole |
| **Auth package (pkg/auth)** | ✅ Complete | JWTManager (access 15min, refresh 7d, HS256), bcrypt password hashing |
| **Token blocklist** | ✅ Complete | Redis-backed JWT blocklist with TTL |
| **Frontend pages** | 🟡 Starter only | Next.js boilerplate page, no app pages yet |
| **Frontend components** | 🔒 Not started | .gitkeep placeholders |
| **Playwright E2E** | 🔒 Not started | Not yet initialized |

---

## Update Protocol

When this file needs updating, use the following rules:

1. **File added**: Add an entry under the correct directory section with a description
2. **File removed**: Remove the entry and note it was deleted
3. **File renamed**: Update the path, keep the description
4. **Directory added**: Add a new subsection if significant, or a line if minor
5. **.gitkeep → real file**: Change 🔒 to 📄 and add a description of what was implemented
6. **Implementation status changed**: Update the Status table at the bottom
