# Rebrand Guimba-GO & Create Missing Skills/Agents

## Overview
Two tasks in one pass:
1. **Rebrand**: Remove all traces of "SPMIS" / "Social Protection Management Information System" and replace with "Guimba" / "Guimba-GO" — this is its own system, not an SPMIS fork.
2. **Extend**: Create the 8 recommended skills and 1 agent identified in the gap analysis.

---

## Part 1: Rebrand — SPMIS → Guimba

### Naming Convention
| Old | New |
|---|---|
| `SPMIS-GO` | `Guimba-GO` |
| `SPMIS` | `Guimba` |
| `Social Protection Management Information System` | `Guimba Batangan Debt Management System` |
| `spmis` (lowercase, in DB/env/containers) | `guimba` |
| `spmis_db` | `guimba_db` |
| `spmis_secret` | `guimba_secret` |
| `spmis-postgres`, `spmis-mongodb`, etc. | `guimba-postgres`, `guimba-mongodb`, etc. |
| `SPMIS API` (Swagger title) | `Guimba API` |
| `test@spmis.gov` | `test@guimba.gov` |
| Go module `github.com/ISubamariner/SPMIS-GO/backend` | `github.com/ISubamariner/guimba-go/backend` |

### Files That Need Changes (exhaustive list)

#### Root Files
- **`AGENTS.md`** — Title line `# SPMIS-GO` → `# Guimba-GO`; description line remove "refactored from C#/.NET" framing, describe as a standalone Go + Next.js system
- **`MASTERPLAN.md`** — All occurrences of `SPMIS`, `SPMIS-GO`, `C# SPMIS`; update title, folder structure labels, phase descriptions. Remove "refactoring from C#" framing — this is a greenfield project
- **`docker-compose.yml`** — Container names (`spmis-postgres` → `guimba-postgres`, etc.), default env values (`spmis` → `guimba`, `spmis_secret` → `guimba_secret`, `spmis_db` → `guimba_db`)
- **`.env.example`** — All `spmis` references in credentials, DB names, DSN strings

#### Backend
- **`backend/go.mod`** — Module path `github.com/ISubamariner/SPMIS-GO/backend` → `github.com/ISubamariner/guimba-go/backend`

#### .github/
- **`.github/copilot-instructions.md`** — Title, project overview paragraph (remove SPMIS identity and C#/.NET refactoring framing)
- **`.github/skills/swagger-gen/SKILL.md`** — Swagger example annotations (`@title SPMIS API`, `@description`)
- **`.github/skills/docker-compose-services/SKILL.md`** — psql connection example (`-U spmis -d spmis_db`), mongosh example
- **`.github/skills/docker-compose-services/references/compose-patterns.md`** — health check example
- **`.github/skills/playwright-testing/SKILL.md`** — `test@spmis.gov` email

#### Documentation/
- **`documentation/README.md`** — Title `# SPMIS-GO Documentation Hub`
- **`documentation/prompts/design-decisions.md`** — Reference to `SPMIS-GO project`
- **`documentation/project/setup-guide.md`** — `cd SPMIS-GO` directory reference

### Rebrand Rules
- Do NOT rename the actual workspace folder on disk (it's already `Guimba-GO`)
- Do NOT rename the GitHub repo slug (`guimba-go` is already correct)
- Do NOT change any Go package names (only the module path in go.mod)
- Replace ALL occurrences — use search to verify zero `spmis`/`SPMIS` matches remain after changes
- When removing the "refactored from C#" framing, replace with: "A full-stack municipal management system built with Go + Next.js"

---

## Part 2: Create Missing Skills

### Skill 1: `auth-rbac`
**Location**: `.github/skills/auth-rbac/SKILL.md`

**Description**: `"Manages authentication (JWT), authorization (RBAC), login/register flows, password hashing, middleware guards, and token refresh. Use when user says 'add auth', 'create login', 'protect route', 'add role', 'JWT', 'middleware guard', 'password hash', 'token refresh', 'sign up', 'register', or when working with middleware/auth*, usecase/user/, or any protected endpoint."`

**SKILL.md body should cover**:
- JWT token generation & refresh flow (access + refresh tokens)
- Password hashing with bcrypt (via `golang.org/x/crypto/bcrypt`)
- Auth middleware pattern for Chi (`middleware/auth.go`)
- Role-based access control: role check middleware, permission model
- Login/Register use case flow through clean architecture layers
- Token storage strategy (httpOnly cookie vs. header)
- Frontend auth: storing tokens, injecting auth headers, handling 401 redirects
- Security rules: never log tokens, never store plaintext passwords, always validate token expiry

**Reference file**: `.github/skills/auth-rbac/references/auth-patterns.md`
- JWT claims structure
- Middleware chain example
- Role hierarchy example (admin > manager > staff > viewer)
- Refresh token rotation example
- Protected route registration example with Chi

### Skill 2: `api-client`
**Location**: `.github/skills/api-client/SKILL.md`

**Description**: `"Manages the centralized frontend API client (src/lib/api.ts), request/response typing, auth header injection, error handling, and token refresh. Use when user says 'API call', 'fetch data', 'api client', 'create api function', 'handle API error', 'token refresh', or when working with src/lib/api.ts or src/types/."`

**SKILL.md body should cover**:
- Base API client structure (`src/lib/api.ts`)
- Typed request/response helpers (generic `apiGet<T>`, `apiPost<T>`, etc.)
- Auth header injection (Bearer token from cookie/storage)
- Automatic 401 handling → redirect to login or trigger token refresh
- Error mapping from backend structured error format to frontend-friendly format
- Request/response interceptor pattern
- Type definitions in `src/types/` matching backend DTOs
- Environment-based base URL (`NEXT_PUBLIC_API_URL`)

**Reference file**: `.github/skills/api-client/references/client-patterns.md`
- Full `api.ts` template
- Error type definitions
- Example typed API function
- Pagination response type

### Skill 3: `error-handling`
**Location**: `.github/skills/error-handling/SKILL.md`

**Description**: `"Standardizes error handling across all Go layers using pkg/apperror/. Use when user says 'handle error', 'error response', 'error code', 'custom error', 'error mapping', 'status code', or when working with pkg/apperror/, handler files, or use case files."`

**SKILL.md body should cover**:
- Error code taxonomy (VALIDATION_ERROR, NOT_FOUND, UNAUTHORIZED, FORBIDDEN, CONFLICT, INTERNAL_ERROR)
- `pkg/apperror/` types: `AppError` struct, constructor functions (`NewNotFound`, `NewValidation`, etc.)
- How handlers map `AppError` → HTTP status code + structured JSON response
- How use cases return domain errors (never HTTP concepts)
- How infrastructure wraps DB errors into domain errors
- Error propagation flow: infrastructure → usecase → handler → HTTP response
- Validation error details array structure
- Rules: never expose internal errors to client, always log full error server-side

**Reference file**: `.github/skills/error-handling/references/error-codes.md`
- Complete error code table with HTTP status mappings
- Example error responses for each code
- Handler helper function template

### Skill 4: `ci-cd`
**Location**: `.github/skills/ci-cd/SKILL.md`

**Description**: `"Manages CI/CD pipelines via GitHub Actions for linting, testing, building, and deploying. Use when user says 'add CI', 'create pipeline', 'GitHub Actions', 'automate tests', 'deploy', 'build pipeline', 'continuous integration', or when working with .github/workflows/."`

**SKILL.md body should cover**:
- GitHub Actions workflow structure
- Pipeline stages: lint → test → build → deploy
- Go pipeline: `golangci-lint`, `go test ./...`, `go build`, Docker image build
- Frontend pipeline: `eslint`, `tsc --noEmit`, `npm test`, `npm run build`
- Migration pipeline: run pending migrations before deploy
- Docker image build & push (to GHCR or Docker Hub)
- Environment strategy: dev (on push to develop), staging (on PR to main), prod (on merge to main)
- Secrets management in GitHub Actions
- Caching strategies (Go modules, npm, Docker layers)

**Reference file**: `.github/skills/ci-cd/references/workflow-templates.md`
- Template: Go CI workflow
- Template: Frontend CI workflow
- Template: Docker build & push workflow
- Template: Migration workflow

### Skill 5: `security-hardening`
**Location**: `.github/skills/security-hardening/SKILL.md`

**Description**: `"Enforces security best practices: CORS, CSP, rate limiting, input sanitization, secure headers, and OWASP Top 10 prevention. Use when user says 'add CORS', 'secure headers', 'rate limit', 'CSP', 'security', 'harden', 'OWASP', 'XSS', 'CSRF', 'injection', or when working with middleware/."`

**SKILL.md body should cover**:
- CORS configuration for Chi (allowed origins, methods, headers)
- Security headers middleware (X-Content-Type-Options, X-Frame-Options, CSP, HSTS, Referrer-Policy)
- Rate limiting middleware (per-IP, per-user, per-endpoint)
- Input validation at delivery layer (already via go-playground/validator, but emphasize SQL injection prevention)
- CSRF protection strategy for cookie-based auth
- Content Security Policy for Next.js
- Request size limits
- OWASP Top 10 checklist mapped to project patterns
- Rules: never trust client input, always validate at boundary, never expose stack traces

**Reference file**: `.github/skills/security-hardening/references/security-checklist.md`
- OWASP Top 10 mapped to project-specific mitigations
- Middleware stack order (security headers → CORS → rate limit → auth → handler)
- CSP template for Next.js

### Skill 6: `redis-caching`
**Location**: `.github/skills/redis-caching/SKILL.md`

**Description**: `"Manages Redis caching patterns: cache-aside, TTL strategy, key naming, invalidation, and session storage. Use when user says 'add cache', 'cache this', 'Redis', 'invalidate cache', 'cache strategy', 'TTL', 'session store', or when working with infrastructure/cache/."`

**SKILL.md body should cover**:
- Cache-aside pattern implementation in Go
- Key naming convention: `{entity}:{id}` for single, `{entity}:list:{hash}` for collections
- TTL strategy: short (1-5min) for lists, medium (15-30min) for single entities, long (1-24h) for config
- Cache invalidation: invalidate on write (create/update/delete)
- Redis client setup in `infrastructure/cache/`
- Cache repository wrapper pattern (wraps real repo, checks cache first)
- Session storage for auth tokens
- Serialization: JSON for complex objects, raw for simple values
- Error handling: cache miss is not an error, cache failure falls through to DB silently
- Connection pooling and health checks

**Reference file**: `.github/skills/redis-caching/references/cache-patterns.md`
- Cache wrapper repository template
- Key naming examples
- TTL cheat sheet
- Invalidation patterns

### Skill 7: `seed-data`
**Location**: `.github/skills/seed-data/SKILL.md`

**Description**: `"Manages database seed data and test fixtures for development and testing. Use when user says 'seed database', 'add test data', 'create fixtures', 'populate database', 'sample data', 'reset and seed', or when working with tests/fixtures/ or migrations/."`

**SKILL.md body should cover**:
- Seed file location: `tests/fixtures/`
- Seed data format: SQL files for Postgres, JSON files for MongoDB
- Seed script: `tests/helpers/seed.go` that loads and executes fixtures
- Development seeds vs. test seeds (different data, same mechanism)
- Idempotent seeds (use `ON CONFLICT DO NOTHING` or upsert)
- Seed data for each entity: realistic but not real data
- Integration with Docker Compose: seed on `docker compose up`
- Integration with tests: seed before test suite, clean after
- Password hashing in seed data (pre-hashed bcrypt values, never plaintext)

**Reference file**: `.github/skills/seed-data/references/fixture-templates.md`
- Example SQL seed file (users, roles)
- Example JSON seed file (audit logs)
- Seed runner Go code template

### Skill 8: `env-config`
**Location**: `.github/skills/env-config/SKILL.md`

**Description**: `"Manages environment configuration, .env files, config validation, and secret handling across environments. Use when user says 'add env variable', 'configure', 'environment', '.env', 'config', 'secret', 'production config', or when working with infrastructure/config/ or .env files."`

**SKILL.md body should cover**:
- Config struct in `infrastructure/config/config.go` with env tags
- Config loading: `.env` file → environment variables → defaults
- Config validation on startup (fail fast if required vars missing)
- `.env.example` as documentation (always kept up to date)
- Secret handling: never commit `.env`, never log secrets, use `*****` in debug output
- Environment hierarchy: `.env.development`, `.env.test`, `.env.production`
- Frontend env: `NEXT_PUBLIC_*` prefix for client-exposed vars
- Docker Compose: env_file directive and variable substitution
- Rules: every new service connection needs a config entry, every config entry needs an `.env.example` entry

**Reference file**: `.github/skills/env-config/references/config-template.md`
- Go config struct template
- Config loader function template
- `.env.example` annotated template

---

## Part 3: Create Missing Agent

### Agent: `feature-orchestrator`
**Location**: `.github/agents/feature-orchestrator.agent.md`

**Description**: `"Orchestrates complete vertical feature slices by coordinating api-builder, db-migrator, and frontend-builder workflows. Use when user says 'add feature', 'create module', 'build CRUD for', 'scaffold feature', 'new resource end-to-end', or when a task spans backend + database + frontend."`

**Agent body should cover**:
- Step 1: **Gather requirements** — entity name, fields, relationships, CRUD operations, which DB (Postgres vs Mongo), auth requirements
- Step 2: **Database layer** — invoke db-migrator pattern: create migration files
- Step 3: **Backend layers** — invoke api-builder pattern: domain entity → repository interface → use case → infrastructure repo → DTOs → handler → routes
- Step 4: **Auth integration** — add middleware guards if endpoint is protected, add role checks if RBAC needed
- Step 5: **Frontend** — invoke frontend-builder pattern: types → API client functions → page components → form components
- Step 6: **Tests** — unit tests for use cases, integration tests for repos, E2E test spec for Playwright
- Step 7: **Documentation** — Swagger annotations on handlers, update README if needed
- Step 8: **Verification checklist** — all layers connected, types match, routes registered, tests pass

---

## Part 4: Update doc-sync Registry

After all skills/agents are created, update `.github/skills/doc-sync/SKILL.md` to add the new files to the documentation registry:

**Tier 3 additions** (agents):
- `.github/agents/feature-orchestrator.agent.md`

**Tier 4 additions** (skills):
- `.github/skills/auth-rbac/SKILL.md`
- `.github/skills/api-client/SKILL.md`
- `.github/skills/error-handling/SKILL.md`
- `.github/skills/ci-cd/SKILL.md`
- `.github/skills/security-hardening/SKILL.md`
- `.github/skills/redis-caching/SKILL.md`
- `.github/skills/seed-data/SKILL.md`
- `.github/skills/env-config/SKILL.md`

**Tier 5 additions** (reference files):
- All `references/*.md` files for the new skills above

---

## Part 5: Update copilot-instructions.md

Add the new skills and agents to the project overview in `.github/copilot-instructions.md` so Copilot knows they exist.

---

## Execution Order
1. Rebrand all files (Part 1) — do all replacements in one pass
2. Create skill files (Part 2) — 8 skills × (SKILL.md + references/*.md) = 16 files
3. Create agent file (Part 3) — 1 file
4. Update doc-sync registry (Part 4)
5. Update copilot-instructions (Part 5)
6. Verify: `grep -ri "spmis" .` returns zero matches (excluding `.git/`)
