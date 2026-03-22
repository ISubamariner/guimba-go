# Conventions

## Naming

| Context | Style | Examples |
|:---|:---|:---|
| Go unexported | `camelCase` | `programService`, `getByID` |
| Go exported | `PascalCase` | `ProgramService`, `GetByID` |
| Go files | `snake_case.go` | `program_service.go`, `user_handler.go` |
| DB columns | `snake_case` | `created_at`, `first_name` |
| DB tables | `snake_case`, plural | `social_programs`, `user_roles` |
| TypeScript vars/functions | `camelCase` | `fetchPrograms`, `userId` |
| TypeScript types/components | `PascalCase` | `ProgramList`, `UserProfile` |
| TypeScript files | `kebab-case.tsx` | `program-list.tsx`, `user-profile.tsx` |
| API routes | `kebab-case` | `/api/v1/social-programs` |
| CSS tokens | `--kebab-case` | `--color-primary`, `--radius-md` |
| Skill folders | `kebab-case` | `bug-tracker`, `design-system` |

## API Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": []
  }
}
```

## Commit Messages
[Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` — new feature
- `fix:` — bug fix
- `refactor:` — code restructuring (no behavior change)
- `docs:` — documentation changes
- `test:` — adding or updating tests
- `chore:` — tooling, deps, config changes

## Branch Strategy
- `master` — production-ready
- `develop` — integration branch
- `feat/<name>` — feature branches
- `fix/<name>` — bugfix branches
