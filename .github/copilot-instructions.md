# Guimba-GO Project Instructions

## Project Overview
Guimba-GO is a full-stack municipal management system built with a **Go backend + Next.js frontend** architecture for the Guimba Batangan Debt Management System.

## Tech Stack
- **Backend**: Go 1.24+, Chi router, pgx (PostgreSQL), mongo-go-driver (MongoDB), go-redis
- **Frontend**: Next.js 15+, TypeScript, Tailwind CSS
- **Database**: PostgreSQL 16+ (relational data), MongoDB 7+ (documents, audit logs, CQRS reads) — both via Docker
- **Cache**: Redis (via Docker)
- **Docs**: Swagger via swaggo/swag
- **Migrations**: golang-migrate (SQL files for Postgres)
- **Auth**: JWT (golang-jwt/jwt/v5), bcrypt (golang.org/x/crypto)

## Coding Standards

### Go Backend
- **Clean Architecture**: `domain/` → `usecase/` → `infrastructure/` → `delivery/`
- Domain layer (`internal/domain/`) has ZERO external framework dependencies (exception: `github.com/google/uuid` accepted as ID primitive)
- Use cases depend only on domain interfaces; infrastructure implements them
- Dependency injection wired in `cmd/server/main.go`
- DTOs in `delivery/http/dto/` — never expose domain entities directly in HTTP responses
- Error handling: return errors, don't panic. Use `pkg/apperror/` custom error types.
- Logging: use `slog` (stdlib)
- Validation: use `go-playground/validator` struct tags on DTOs
- Tests: Go tests in `backend/tests/` (unit, integration); Playwright E2E in root `tests/`
- API versioning: all routes prefixed with `/api/v1/`

### Frontend
- Use Next.js App Router (not Pages Router)
- TypeScript strict mode enabled
- Components go in `src/components/`, pages in `src/app/`
- API calls go through a centralized client in `src/lib/api.ts`

### Database
- All schema changes via migration files (never manual DDL)
- Migration files: `{timestamp}_{description}.up.sql` / `.down.sql`
- Use parameterized queries (never string concatenation for SQL)

## Guardrails

### Anti-Redundancy Check
Before generating new code, verify:
1. Does a similar function/handler/component already exist in the codebase?
2. Can an existing utility in `pkg/` or `src/lib/` be reused or extended?
3. Check the skill descriptions — has a skill already been created for this task?

If duplication is found, extend or refactor the existing code instead of creating new files.

### Structured Bug Logging Convention
When a bug is encountered and resolved, log it in `.github/skills/bug-tracker/references/bug-log.md` using this format:

```markdown
### [SHORT_TITLE] — YYYY-MM-DD
- **Issue**: What went wrong (symptoms, error messages)
- **Root Cause**: Why it happened (the actual underlying problem)
- **Resolution**: What was changed to fix it
- **Files Changed**: List of modified files
- **Prevention**: What rule or check would prevent recurrence
```

### Before Debugging
Before attempting to fix any bug:
1. Search `.github/skills/bug-tracker/references/bug-log.md` for **related keywords**
2. Check if the same root cause has appeared before
3. If a match is found, apply the documented resolution pattern first

### Documentation Sync (Long-Term Memory)
After completing any meaningful code change (feature, fix, refactor, new dependency):
1. Consider which documentation files may be affected
2. Use the `doc-sync` skill to update all relevant docs, instructions, and skills
3. Every doc update is logged in `.github/skills/doc-sync/references/changelog.md`

This is the project's long-term memory — stale docs lead to hallucinated guidance in future sessions.

### Iterative Refinement
When updating existing code or skills:
- Do not rewrite from scratch — use targeted, diff-style changes
- Preserve existing working logic
- Only modify what's necessary to address the issue

## MCP Servers (Tool Providers)

This project has 9 MCP servers configured to extend Copilot's capabilities with live data access and specialized tools. Use them proactively — don't guess when you can query.

| Server | What It Provides | When to Use |
|:---|:---|:---|
| `postgres` | Direct SQL queries on PostgreSQL | Verify schema, check data, test queries, inspect migrations — before writing repo code |
| `mongodb` | Read-only MongoDB access | Inspect audit logs, check document schemas, verify CQRS read models |
| `redis` | Redis key/value operations | Check cached data, inspect token blocklist, verify TTLs |
| `memory` | Persistent key-value memory | Store session context, track cross-task state, remember decisions |
| `filesystem` | Read/write project files | Bulk file operations, directory traversal, file metadata |
| `playwright` | Browser automation & testing | Run E2E tests, capture screenshots, validate UI flows |
| `chrome-devtools` | Chrome DevTools Protocol | Inspect network requests, debug frontend, audit performance |
| `context7` | Up-to-date library documentation | Look up current API docs for any npm/Go package instead of guessing |
| `markitdown` | Convert files to Markdown | Convert PDFs, DOCX, XLSX to Markdown for analysis |

### MCP Usage Rules
- **Database servers require Docker running** — run `docker compose up -d` first (see `docker-compose-services` skill)
- **Query before coding** — use `postgres` MCP to verify table schemas before writing repository code
- **Use `context7` for docs** — always check current library APIs instead of relying on training data
- **MongoDB is read-only** — use it for inspection only; writes go through Go application code
- **Memory server** — use for tracking multi-step task progress across conversation turns

## Available Skills & Agents

### Skills (auto-triggered by description matching)
| Skill | Purpose |
|:---|:---|
| `docker-compose-services` | Local dev environment (PostgreSQL, MongoDB, Redis) |
| `swagger-gen` | Swagger/OpenAPI doc generation from Go annotations |
| `go-testing` | Go test patterns, table-driven tests, mocking |
| `bug-tracker` | Persistent bug memory, pattern recognition |
| `doc-sync` | Documentation sync & audit (long-term memory) |
| `playwright-testing` | Browser E2E testing & visual regression |
| `design-system` | Consolidated CSS, design tokens, UI primitives |
| `auth-rbac` | JWT authentication, RBAC, login/register flows |
| `api-client` | Frontend API client, typed requests, error handling |
| `error-handling` | Standardized error codes & propagation via `pkg/apperror/` |
| `ci-cd` | GitHub Actions CI/CD pipelines |
| `security-hardening` | CORS, CSP, rate limiting, OWASP Top 10 |
| `redis-caching` | Cache-aside pattern, TTL strategy, invalidation |
| `seed-data` | Database seed data & test fixtures |
| `env-config` | Environment configuration, .env, secret handling |

### Agents (invoked on demand)
| Agent | Purpose |
|:---|:---|
| `api-builder` | Scaffold backend API endpoints (domain → handler) |
| `frontend-builder` | Scaffold frontend pages & components |
| `db-migrator` | Create database migration files |
| `feature-orchestrator` | Orchestrate full vertical feature slices (DB + backend + frontend) |
