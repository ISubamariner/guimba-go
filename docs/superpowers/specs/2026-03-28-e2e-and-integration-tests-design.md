# E2E & Integration Tests Design

**Goal:** Add comprehensive test coverage across two independent subsystems: Playwright E2E tests for the frontend and Go integration tests for the backend.

**Approach:** Two separate implementation plans, executed sequentially (Playwright first, then backend integration).

---

## Part 1: Playwright E2E Tests

### Goal

Browser-based tests covering critical user flows across all frontend pages. Validates the full stack (frontend + backend + database) from the user's perspective.

### Architecture

- **Playwright** with Chromium as the primary browser
- **Page Object Model** pattern for maintainable selectors and actions
- **Hybrid data strategy:** SQL-seeded admin user (via API register endpoint at global setup) + API-driven test-specific data per test
- Tests organized by feature module in `tests/playwright/specs/`
- Requires backend + frontend running (either `docker compose up` or local dev servers)

### Components

| Component | Path | Purpose |
|---|---|---|
| package.json | `tests/playwright/package.json` | Playwright dependencies |
| Config | `tests/playwright/playwright.config.ts` | Base URL (`http://localhost:3000`), timeouts, Chromium setup, global setup/teardown |
| Global Setup | `tests/playwright/helpers/global-setup.ts` | Register + login admin user via API, save auth state |
| API Helper | `tests/playwright/helpers/api-client.ts` | Typed API calls for creating test data (tenants, properties, debts) |
| Auth Fixture | `tests/playwright/fixtures/auth.fixture.ts` | Extends Playwright `test` with pre-authenticated page (reuses stored auth state) |
| Page Objects | `tests/playwright/pages/*.page.ts` | One per page: login, dashboard, tenants, properties, debts, transactions, audit |
| Specs | `tests/playwright/specs/<module>/*.spec.ts` | Test files grouped by feature |

### Page Objects

Each page object encapsulates selectors and common actions:

| Page Object | Key Methods |
|---|---|
| `LoginPage` | `goto()`, `login(email, password)`, `getError()` |
| `DashboardPage` | `goto()`, `getStatValue(label)`, `getActivities()` |
| `TenantsPage` | `goto()`, `openCreateModal()`, `fillCreateForm(data)`, `submit()`, `getTableRows()` |
| `PropertiesPage` | `goto()`, `openCreateModal()`, `fillCreateForm(data)`, `submit()`, `getTableRows()` |
| `DebtsPage` | `goto()`, `openCreateModal()`, `fillCreateForm(data)`, `submit()`, `openPayModal(row)`, `openCancelModal(row)`, `getStatusBadge(row)` |
| `TransactionsPage` | `goto()`, `getTableRows()`, `getVerifiedBadge(row)` |
| `AuditPage` | `goto()`, `getTableRows()`, `getActionBadge(row)` |

### Test Specs

#### 1. Auth (`specs/auth/`)
- `login.spec.ts` — Successful login redirects to dashboard; invalid credentials show error; empty form shows validation
- `register.spec.ts` — Successful registration redirects to dashboard; duplicate email shows error; password mismatch shows error
- `logout.spec.ts` — Logout clears session, redirects to login
- `auth-guard.spec.ts` — Unauthenticated access to `/dashboard` redirects to `/login`

#### 2. Dashboard (`specs/dashboard/`)
- `dashboard.spec.ts` — Stats cards render with numeric values; recent activities list renders; activity items have timestamps and action badges

#### 3. Tenants (`specs/tenants/`)
- `tenants-crud.spec.ts` — List page renders table; create tenant via modal, verify row appears; table shows name, email, phone, status

#### 4. Properties (`specs/properties/`)
- `properties-crud.spec.ts` — List page renders table; create property via modal with type/size/rent; verify row appears

#### 5. Debts (`specs/debts/`)
- `debts-crud.spec.ts` — Create debt for existing tenant; status shows PENDING badge
- `debts-pay.spec.ts` — Pay a pending debt; status transitions to PAID; transaction appears in transactions page
- `debts-cancel.spec.ts` — Cancel a pending debt with reason; status transitions to CANCELLED

#### 6. Transactions (`specs/transactions/`)
- `transactions.spec.ts` — List renders after payment flow; shows amount, method, type badge, verified badge

#### 7. Audit (`specs/audit/`)
- `audit.spec.ts` — Admin sees audit log entries; entries have timestamp, action, resource type

#### 8. Navigation (`specs/navigation/`)
- `sidebar.spec.ts` — Admin sees all nav items; verify sidebar highlights active route

### Prerequisites

- Backend running at `http://localhost:8080`
- Frontend running at `http://localhost:3000`
- Clean database (tests create their own data)

### Run Commands

```bash
# Install
cd tests/playwright && npm install

# Run all E2E tests
cd tests/playwright && npx playwright test

# Run specific spec
cd tests/playwright && npx playwright test specs/auth/login.spec.ts

# Run with UI mode
cd tests/playwright && npx playwright test --ui

# Update visual snapshots
cd tests/playwright && npx playwright test --update-snapshots
```

---

## Part 2: Backend Integration Tests

### Goal

Test repository implementations and API endpoints against real PostgreSQL and MongoDB using testcontainers-go. Validates that SQL queries, migrations, and full request/response cycles work correctly.

### Architecture

- **testcontainers-go** spins up ephemeral Postgres 16 + MongoDB 7 containers per test suite
- Migrations auto-applied to test Postgres container on startup
- Build tag `//go:build integration` separates from unit tests
- Two test categories: repository tests (DB layer) and API tests (full HTTP stack)

### Components

| Component | Path | Purpose |
|---|---|---|
| Container helpers | `backend/tests/helpers/testcontainers.go` | Start Postgres + Mongo containers, run migrations, return pools/clients |
| Test main | `backend/tests/integration/integration_test.go` | `TestMain` — starts containers once for the entire package, tears down after |
| Repo tests | `backend/tests/integration/*_repo_test.go` | Test each repository implementation against real DB |
| API tests | `backend/tests/integration/*_api_test.go` | Full HTTP: build real server, hit endpoints, assert responses + DB state |
| Test data | `backend/tests/integration/testdata/` | Optional seed SQL for prerequisite data |

### Test Isolation Strategy

- Containers start once per `TestMain` (not per test function — too slow)
- Migrations run once on container startup
- Each test function uses a transaction that rolls back after the test (for Postgres)
- For MongoDB tests, each test uses a unique collection prefix or cleans up after itself

### Repository Tests

| Test File | What It Tests |
|---|---|
| `program_repo_test.go` | Create, GetByID, List (pagination, filtering), Update, Delete |
| `user_repo_test.go` | Create, GetByEmail, GetByID, List, Update, Delete |
| `role_repo_test.go` | AssignRole, RemoveRole, GetUserRoles |
| `beneficiary_repo_test.go` | CRUD + EnrollInProgram, RemoveFromProgram, GetWithPrograms |
| `tenant_repo_test.go` | CRUD (landlord-scoped), Deactivate, list filtering |
| `property_repo_test.go` | CRUD (landlord-scoped), Deactivate, list filtering |
| `debt_repo_test.go` | CRUD, state transitions (UpdateStatus), GetByTenant, overdue detection |
| `transaction_repo_test.go` | Create, GetByID, List, Verify, GetByDebt |
| `audit_repo_test.go` | Log (MongoDB write), Query with filters (admin scope, landlord scope) |

### API Tests

| Test File | What It Tests |
|---|---|
| `auth_api_test.go` | Register → Login → Refresh → Me → Logout flow; invalid credentials; duplicate registration |
| `programs_api_test.go` | CRUD endpoints; public GET vs admin-only POST/PUT/DELETE |
| `beneficiaries_api_test.go` | CRUD + enrollment; role-gated access |
| `tenants_api_test.go` | CRUD + deactivation; landlord scoping |
| `debts_api_test.go` | Create → Pay → verify PAID status; Create → Cancel; partial payment |
| `transactions_api_test.go` | List after payment; verify transaction |

### Test Helper API

```go
// testcontainers.go
type TestContainers struct {
    PgPool  *pgxpool.Pool
    MongoDB *mongo.Database
}

func SetupContainers(t *testing.T) *TestContainers  // starts containers, runs migrations
func (tc *TestContainers) Cleanup()                   // stops containers
func (tc *TestContainers) TruncateAll(t *testing.T)   // clear all tables between tests
```

### Run Commands

```bash
# Run integration tests (requires Docker)
cd backend && go test -tags=integration ./tests/integration/... -v

# Run specific test
cd backend && go test -tags=integration ./tests/integration/... -run TestProgramRepo -v

# Run with race detection
cd backend && go test -tags=integration -race ./tests/integration/... -v
```

---

## Implementation Order

1. **Playwright E2E tests first** — validates the full stack (frontend + backend), higher immediate value since frontend has zero test coverage
2. **Backend integration tests second** — validates DB layer in isolation, catches SQL/query bugs

Each gets its own implementation plan via the writing-plans skill.
