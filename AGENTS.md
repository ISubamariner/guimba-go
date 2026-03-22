# Guimba-GO

## What Is This?
Guimba — a full-stack municipal management system built with **Go + Next.js** for the Guimba Batangan Debt Management System.

## Architecture
```
Client (Next.js) → API Gateway (Go/Chi) → Services → Repositories → PostgreSQL
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
