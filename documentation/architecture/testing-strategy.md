# Testing Strategy

## Test Pyramid

```
        ╱ E2E (Playwright) ╲          ← Few, slow, high confidence
       ╱  Integration Tests  ╲        ← Moderate, real DB
      ╱    Unit Tests (Go)    ╲       ← Many, fast, isolated
     ╱  Frontend Unit (Jest)   ╲      ← Many, fast, isolated
    ╱──────────────────────────╲
```

## Test Types & Locations

| Type | Location | Scope | Dependencies | Speed |
|:---|:---|:---|:---|:---|
| **Go Unit** | `backend/tests/unit/` | Domain entities, use cases, handlers, middleware | Mocks only | Fast (~ms) |
| **Go Integration** | `backend/tests/integration/` | Repository impls, full API calls | Real DB (testcontainers) | Medium (~s) |
| **Go E2E** | `backend/tests/e2e/` | Multi-step API flows | Running server + DB | Slow (~s) |
| **Playwright** | `tests/playwright/specs/` | Browser UI flows, visual regression | Full stack running | Slow (~s) |
| **Frontend Unit** | `frontend/__tests__/` | React components, hooks, lib | Jest/Vitest mocks | Fast (~ms) |

> **Why `backend/tests/` instead of root `tests/`?** Go's `internal` package visibility rules prevent external modules from importing `internal/` packages. Since tests need to import handlers, middleware, and DTOs from `internal/`, they must live within the `backend/` module.

## Shared Resources

| Resource | Location | Used By |
|:---|:---|:---|
| Mocks | `backend/tests/mocks/` | Go unit tests |
| Fixtures (JSON, SQL) | `tests/fixtures/` | Integration, E2E |
| Test helpers (Go) | `backend/tests/helpers/` | Go unit & integration tests |
| Playwright pages | `tests/playwright/pages/` | Playwright specs |
| Playwright fixtures | `tests/playwright/fixtures/` | Playwright specs |
| Visual baselines | `tests/playwright/snapshots/` | Playwright visual regression |

## Running Tests

```bash
# All Go unit tests
cd backend && go test ./tests/unit/...

# All Go integration tests
cd backend && go test -tags=integration ./tests/integration/...

# All Playwright tests
cd tests/playwright && npx playwright test

# Playwright with UI debugger
cd tests/playwright && npx playwright test --ui

# Update visual regression baselines
cd tests/playwright && npx playwright test --update-snapshots

# Frontend unit tests
cd frontend && npm test

# Full coverage report (Go)
cd backend && go test -coverprofile=coverage.out ./tests/... && go tool cover -html=coverage.out
```

## Playwright Specifics

- **Page Object Model**: Every page gets a class in `tests/playwright/pages/`
- **Auth via API**: Login through API in fixtures, not through UI (faster)
- **Visual Regression**: Screenshots stored in `snapshots/`, compared with `toHaveScreenshot()`
- **Full-Stack Validation**: Tests can call API directly + assert in browser + verify DB state
- **3 scope levels**: UI-only specs, API validation specs, visual regression specs

## Current Test Coverage

| Domain | Unit Tests | Integration Tests | E2E Tests |
|:---|:---|:---|:---|
| **Infrastructure** | 9 tests (apperror, middleware, health DTO) | — | — |
| **Programs** | 26 tests (entity validation, 5 use cases, handler CRUD) | — (pending) | — |
| **Users & Auth** | 22 tests (register, login, refresh, profile, logout, JWT, blocklist, user CRUD, role assignment) | — (pending) | — |
| **Beneficiaries** | 31 tests (entity validation, 7 use cases incl. enrollment, handler CRUD + enrollment) | — (pending) | — |

## Conventions

- Every use case gets at least one unit test (table-driven)
- Every repository implementation gets an integration test
- Every critical user flow gets a Playwright spec
- Key pages get visual regression baselines
- Tests are independent — no shared mutable state between tests
- Always check `bug-log.md` before debugging a test failure
