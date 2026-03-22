# Documentation Changelog

Record of all documentation updates performed by the doc-sync skill. Newest entries first.

---

<!-- Format:
### YYYY-MM-DD — [TRIGGER_DESCRIPTION]
- **Trigger**: What code change prompted this update
- **Files Updated**: List of docs modified
- **Changes**: Brief description of what was updated in each file
-->

### 2026-03-22 — MCP Server Documentation & Optimization
- **Trigger**: 9 MCP servers configured (postgres, mongodb, redis, memory, filesystem, playwright, chrome-devtools, context7, markitdown) — documentation needed to optimize usage
- **Files Updated**: .github/copilot-instructions.md, AGENTS.md, .github/instructions/database.instructions.md, .github/instructions/go-backend.instructions.md, .github/instructions/nextjs-frontend.instructions.md, documentation/copilot-config/README.md, documentation/copilot-config/mcp-servers.md (NEW), .vscode/mcp.json, changelog.md
- **Changes**:
  - **copilot-instructions.md**: Added "MCP Servers" section with server table, usage rules
  - **AGENTS.md**: Added "MCP Servers" section with compact reference
  - **database.instructions.md**: Added "MCP Server Integration" section
  - **go-backend.instructions.md**: Added "MCP-Assisted Development" section
  - **nextjs-frontend.instructions.md**: Added "MCP-Assisted Development" section
  - **copilot-config/README.md**: Added MCP config locations and server list
  - **copilot-config/mcp-servers.md**: NEW — full MCP reference (capabilities, usage patterns, troubleshooting)
  - **.vscode/mcp.json**: Added missing chrome-devtools and markitdown servers

### 2026-03-22 — Phase 4 Beneficiaries Module Complete
- **Trigger**: Phase 4 implementation — Beneficiaries with program enrollment (entity, 7 use cases, PG repo, DTOs, handler with Swagger, many-to-many program_beneficiaries junction, 31 new tests, 88 total)
- **Files Updated**: file-tree.md, clean-architecture.md, testing-strategy.md, api/README.md, auth-rbac/references/auth-patterns.md, error-handling/references/error-codes.md, seed-data/SKILL.md, changelog.md
- **Changes**:
  - **file-tree.md**: Added 17 new files (migration, entity, repository, 7 use cases, PG repo, DTOs, handler, mock, 3 test files); updated key backend files table; updated implementation status table (entities/repos/use cases/persistence/handlers/DTOs/mocks/tests all now include Beneficiaries; 88 total tests)
  - **clean-architecture.md**: Extended DI wiring with Beneficiary module injection (repo → 7 use cases → handler)
  - **testing-strategy.md**: Added Beneficiaries row to coverage table (31 tests)
  - **api/README.md**: Added Beneficiaries endpoint table (7 endpoints: CRUD + enroll/remove from program) + query parameters table
  - **auth-patterns.md**: Added beneficiary routes to staff+ middleware chain example
  - **error-codes.md**: Added BAD_REQUEST code row (was missing from table)
  - **seed-data/SKILL.md**: Updated fixture file listing (replaced borrowers/debts with beneficiaries/programs/program_beneficiaries)

### 2026-03-22 — Phase 4 Users & Auth Module Complete
- **Trigger**: Phase 4 implementation — Users & Auth with full RBAC (JWT, bcrypt, Redis blocklist, role-based middleware, 6 migrations, 22 new unit tests)
- **Files Updated**: copilot-instructions.md, tech-stack.md, auth-rbac/SKILL.md, auth-rbac/references/auth-patterns.md, redis-caching/SKILL.md, seed-data/SKILL.md, file-tree.md, clean-architecture.md, testing-strategy.md, api/README.md, changelog.md
- **Changes**:
  - **copilot-instructions.md**: Updated Auth line to `golang-jwt/jwt/v5` + `golang.org/x/crypto`
  - **tech-stack.md**: Split Auth row into jwt/v5 + bcrypt separate entries
  - **auth-rbac/SKILL.md**: Fixed role hierarchy (3 roles not 4: admin/staff/viewer, no manager), updated Claims struct (uuid.UUID), added TokenBlocklist section, updated middleware signatures
  - **auth-rbac/references/auth-patterns.md**: Fixed role hierarchy, updated Claims struct, replaced stale middleware chain example with actual routes (programs write=staff+admin, users=admin), removed manager references
  - **redis-caching/SKILL.md**: Added `token_blocklist:{jti}` key pattern, added token-bound TTL row
  - **seed-data/SKILL.md**: Fixed role seeds (removed manager), added note about migration-based seeding
  - **file-tree.md**: Major update — added 30+ new files (6 migrations, user/role entities, user/role repos, jwt.go, password.go, token_blocklist.go, user/role PG repos, auth middleware, auth/user DTOs, auth/user handlers, auth/user use cases, 3 mocks, 2 test files); updated key backend files table; updated implementation status table
  - **clean-architecture.md**: Extended DI wiring example with Auth + User module injection (JWTManager, TokenBlocklist, repos, use cases, handlers)
  - **testing-strategy.md**: Added Users & Auth row to coverage table (22 tests)
  - **api/README.md**: Added Auth endpoints table (register, login, refresh, me, logout) and Users endpoints table (list, update, delete, assign role) with Auth column

### 2026-03-22 — Phase 4 Programs Module Complete
- **Trigger**: Phase 4 implementation — Programs domain module (entity, repository, 5 use cases, PG persistence, DTOs, handler with Swagger, router, 35 unit tests)
- **Files Updated**: MASTERPLAN.md, documentation/architecture/clean-architecture.md, documentation/architecture/testing-strategy.md, documentation/api/README.md, .github/skills/doc-sync/references/file-tree.md, .github/skills/doc-sync/references/changelog.md
- **Changes**:
  - **MASTERPLAN.md**: Checked off Programs in Phase 4 port order with completion summary
  - **clean-architecture.md**: Replaced "Phase 4 will add" placeholder with actual Programs wiring code (repo → use cases → handler → router.Handlers struct)
  - **testing-strategy.md**: Added "Current Test Coverage" table showing 9 infrastructure + 26 Programs tests
  - **api/README.md**: Added "Available Endpoints" section with Programs CRUD routes and query parameters table
  - **file-tree.md**: Major update — replaced 8 "empty/awaiting Phase 4" entries with actual file descriptions (entity, errors, repository interface, 5 use cases, PG repo, DTOs, handler, router); added mocks and 3 new test files; updated implementation status table (9 items changed from 🔒 to ✅); updated key backend files table with 6 new entries; updated main.go description

### 2026-03-22 — Phase 3 Core Backend Setup Complete
- **Trigger**: Phase 3 implementation — config, database connections, cache, HTTP layer, migrations, swagger, tests, main.go wiring
- **Files Updated**: MASTERPLAN.md, .github/copilot-instructions.md, AGENTS.md, .github/instructions/go-backend.instructions.md, documentation/architecture/testing-strategy.md, documentation/architecture/clean-architecture.md, documentation/project/setup-guide.md, .github/skills/env-config/SKILL.md, .github/skills/doc-sync/references/file-tree.md, .github/skills/doc-sync/references/changelog.md
- **Changes**:
  - **MASTERPLAN.md**: Checked off 12 of 13 Phase 3 tasks (persistence implementations deferred to Phase 4); added note about Go test location
  - **copilot-instructions.md**: Updated test location from "centralized `tests/`" to "Go tests in `backend/tests/`; Playwright E2E in root `tests/`"
  - **AGENTS.md**: Added MongoDB to architecture diagram
  - **go-backend.instructions.md**: Updated testing section — paths changed from `tests/` to `backend/tests/`, added note about internal package visibility, added Playwright location
  - **testing-strategy.md**: Updated all Go test paths to `backend/tests/`, added explanation for why tests moved, updated shared resources table, updated running tests commands
  - **clean-architecture.md**: Replaced pseudo-code main.go example with actual implemented wiring (Viper config, pgx pool, mongo client, redis client, migration runner, Chi router)
  - **setup-guide.md**: Fixed Redis port (6380 external), updated migration section (auto-run on startup), fixed integration test command path
  - **env-config/SKILL.md**: Replaced godotenv/envconfig examples with actual Viper-based implementation (Config struct with mapstructure tags, Viper loader, DSN auto-construction)
  - **file-tree.md**: Major update — replaced all .gitkeep entries in infrastructure/delivery/docs/migrations with actual file descriptions; added backend/tests/ section; updated implementation status table (12 items now ✅); added note about root tests/ being for Playwright only

### 2026-03-22 — Project File Tree & File Registry Created
- **Trigger**: Need for a running inventory of all project files with descriptions for orientation and staleness detection
- **Files Updated**: .github/skills/doc-sync/references/file-tree.md (NEW), .github/skills/doc-sync/SKILL.md, .github/skills/doc-sync/references/changelog.md
- **Changes**:
  - Created `references/file-tree.md` — complete file tree of every file in the project with icons, descriptions, implementation status table, and update protocol
  - Updated `SKILL.md` — added file-tree.md to Tier 5 registry and business-logic-reference.md to Tier 6 registry

### 2026-03-22 — Business Logic Reference Extracted from guimba-debt-tracker
- **Trigger**: Full extraction of business logic from original Python/FastAPI project (guimba-debt-tracker v3.1.0) for behavioral parity during Go rewrite
- **Files Updated**: documentation/prompts/business-logic-reference.md (NEW), documentation/README.md, MASTERPLAN.md, .github/skills/doc-sync/references/changelog.md
- **Changes**:
  - Created `documentation/prompts/business-logic-reference.md` — comprehensive reference covering all 8 domain entities, 3 value objects, 10+ business rule sets, 10 service workflows, auth flows, background jobs, dashboard stats, OCR, audit system, error taxonomy, and 62+ API endpoints
  - Updated `documentation/README.md` — added business-logic-reference.md to directory map and quick links table
  - Updated `MASTERPLAN.md` — added "Business Logic Reference" section before Phase 1 with summary and usage rule

### 2026-03-19 — MongoDB Added (Polyglot Persistence)
- **Trigger**: Decision to add MongoDB alongside PostgreSQL for flexible per-module database choice
- **Files Updated**: MASTERPLAN.md, copilot-instructions.md, go-backend.instructions.md, database.instructions.md, api-builder.agent.md, docker-compose-services/SKILL.md, compose-patterns.md, documentation/project/tech-stack.md, documentation/project/setup-guide.md, documentation/architecture/clean-architecture.md
- **Changes**:
  - MASTERPLAN: Added MongoDB to environment snapshot, Phase 1, Phase 2 (polyglot persistence strategy table, persistence/pg/ + persistence/mongo/ split), Phase 3 tasks, Key Libraries
  - copilot-instructions.md: Updated tech stack with MongoDB
  - go-backend.instructions.md: Infrastructure layer now documents pg/ and mongo/ subdirectories
  - database.instructions.md: Added polyglot persistence section, MongoDB collection/query standards, DB selection guide
  - api-builder.agent.md: Step 4 now offers PostgreSQL or MongoDB per resource
  - docker-compose-services: Added MongoDB service, connection command, health check
  - documentation/: Updated tech-stack.md, setup-guide.md, clean-architecture.md with polyglot persistence example

### 2026-03-19 — Documentation Hub Created
- **Trigger**: Decision to consolidate all system docs, prompts, and Copilot config mirrors into `documentation/`
- **Files Updated**: MASTERPLAN.md, doc-sync/SKILL.md, new `documentation/` directory (11 files)
- **Changes**:
  - MASTERPLAN: Added `documentation/` to folder structure with 6 subdirectories
  - doc-sync/SKILL.md: Added Tier 6 (documentation hub) with 9 files to the registry; added `documentation/README.md` to Tier 1
  - Created: documentation/README.md (index), project/setup-guide.md, project/tech-stack.md, project/conventions.md, architecture/clean-architecture.md, architecture/testing-strategy.md, architecture/design-system.md, api/README.md, prompts/connected-trio-original.md, prompts/design-decisions.md, copilot-config/README.md

### 2026-03-19 — Consolidated Design System Added
- **Trigger**: Decision to enforce UI consistency with consolidated CSS and design tokens
- **Files Updated**: MASTERPLAN.md, nextjs-frontend.instructions.md, doc-sync/SKILL.md, new skill design-system/
- **Changes**:
  - MASTERPLAN: Expanded frontend folder structure with `src/styles/` (tokens, typography, layouts, components), `src/components/ui/` primitives, `tailwind.config.ts`; Phase 5 tasks expanded with design system foundation + UI primitives
  - nextjs-frontend.instructions.md: Replaced simple styling section with full design system architecture (token flow, rules, component hierarchy, violation examples)
  - doc-sync/SKILL.md: Added design-system to Tier 4 and Tier 5 registries
  - New skill: design-system/ with SKILL.md (enforcement rules, violation detection, audit flow, component patterns, tokens↔tailwind connection) and references/token-registry.md (full token inventory, component CSS classes, primitive component list)

### 2026-03-19 — Playwright E2E + Visual Regression Added
- **Trigger**: Decision to add Playwright for full-stack browser E2E and visual regression testing
- **Files Updated**: MASTERPLAN.md, doc-sync/SKILL.md, new skill playwright-testing/
- **Changes**:
  - MASTERPLAN: Added `tests/playwright/` to folder structure with specs/, pages/, fixtures/, helpers/, snapshots/; updated testing strategy table; added Playwright commands; added to Phase 2 tasks and Phase 5 tasks; added to Key Libraries
  - doc-sync/SKILL.md: Added playwright-testing to Tier 4 and Tier 5 registries
  - New skill: playwright-testing/ with SKILL.md (POM, fixtures, full-stack validation, visual regression, troubleshooting) and references/playwright-patterns.md (config, CI, locator strategy, tagging)

### 2026-03-19 — Clean Architecture + Centralized Tests Adoption
- **Trigger**: Architecture decision to adopt Clean Architecture and separate tests folder
- **Files Updated**: MASTERPLAN.md, copilot-instructions.md, go-backend.instructions.md, api-builder.agent.md, go-testing/SKILL.md
- **Changes**: 
  - MASTERPLAN Phase 2: replaced flat `internal/` layout with Clean Architecture layers (domain → usecase → infrastructure → delivery); added `tests/` folder structure (unit/integration/e2e/fixtures/mocks/helpers)
  - MASTERPLAN Phase 3 & 4: updated tasks to reference Clean Architecture paths
  - copilot-instructions.md: updated Go backend section with Clean Architecture rules
  - go-backend.instructions.md: full rewrite to document all 4 layers with dependency rules
  - api-builder.agent.md: updated scaffolding workflow to generate files layer-by-layer (domain first, delivery last)
  - go-testing/SKILL.md: updated test paths to `tests/`, mock paths to `tests/mocks/`, run commands to `./tests/...`

### 2026-03-19 — Initial Setup
- **Trigger**: Phase 0 completion — Copilot customization layer created
- **Files Updated**: All (initial creation)
- **Changes**: Created full documentation layer: copilot-instructions.md, AGENTS.md, 3 path instructions, 3 agents, 5 skills (docker-compose-services, swagger-gen, go-testing, bug-tracker, doc-sync)
