# Guimba-GO

## What Is This?
Guimba — a full-stack municipal management system built with **Go + Next.js** for the Guimba Batangan Debt Management System.

## Architecture
```
Client (Next.js) → API Gateway (Go/Chi) → Services → Repositories → PostgreSQL
                                                                   → MongoDB (audit, docs)
                                                                   → Redis (cache)
```

## Conventions

### Naming
- Go: `camelCase` for unexported, `PascalCase` for exported, `snake_case` for DB columns
- TypeScript: `camelCase` for variables/functions, `PascalCase` for types/components
- Files: `snake_case.go`, `kebab-case.tsx`
- API routes: `kebab-case` (`/api/v1/social-programs`)

### Error Responses
All API errors use this shape:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": []
  }
}
```

### Commit Messages
Use Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`

### Branch Strategy
- `main` — production-ready
- `develop` — integration branch
- `feat/<name>` — feature branches
- `fix/<name>` — bugfix branches

## MCP Servers

9 MCP servers are configured for direct data access and tooling. Prefer querying live data over guessing.

| Server | Purpose |
|:---|:---|
| `postgres` | SQL queries against PostgreSQL (schema verification, data checks) |
| `mongodb` | Read-only MongoDB access (audit logs, documents) |
| `redis` | Redis key/value ops (cache inspection, token blocklist) |
| `memory` | Persistent key-value store for session context |
| `filesystem` | Project file read/write |
| `playwright` | Browser automation & E2E testing |
| `chrome-devtools` | Chrome DevTools Protocol (network, performance) |
| `context7` | Current library documentation lookup |
| `markitdown` | Convert files (PDF, DOCX) to Markdown |

**Rule**: Query `postgres` to verify schemas before writing Go repository code. Use `context7` for up-to-date library APIs.
