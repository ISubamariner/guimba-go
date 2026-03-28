# Backend Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Go integration tests for all repository implementations and key API endpoints using testcontainers-go against real PostgreSQL and MongoDB.

**Architecture:** testcontainers-go spins up ephemeral Postgres 16 + MongoDB 7 containers. Migrations auto-applied on startup. Build tag `//go:build integration` separates from unit tests. Each test uses a clean database state via `TruncateAll()` between tests.

**Tech Stack:** Go, testcontainers-go, pgx v5, mongo-go-driver v2, golang-migrate v4

**Prerequisites:** Docker running (testcontainers needs it).

---

## File Structure

| File | Purpose |
|---|---|
| `backend/tests/helpers/testcontainers.go` | Start/stop Postgres + Mongo containers, run migrations, provide cleanup |
| `backend/tests/integration/integration_test.go` | `TestMain` — starts containers once, tears down after all tests |
| `backend/tests/integration/program_repo_test.go` | Program repository integration tests |
| `backend/tests/integration/user_repo_test.go` | User repository integration tests |
| `backend/tests/integration/tenant_repo_test.go` | Tenant repository integration tests |
| `backend/tests/integration/property_repo_test.go` | Property repository integration tests |
| `backend/tests/integration/debt_repo_test.go` | Debt repository integration tests |
| `backend/tests/integration/transaction_repo_test.go` | Transaction repository integration tests |
| `backend/tests/integration/beneficiary_repo_test.go` | Beneficiary repository integration tests |
| `backend/tests/integration/audit_repo_test.go` | Audit repository (MongoDB) integration tests |
| `backend/tests/integration/auth_api_test.go` | Auth API endpoint integration tests |

---

### Task 1: Testcontainers Helper

**Files:**
- Create: `backend/tests/helpers/testcontainers.go`

- [ ] **Step 1: Add testcontainers-go dependency**

```bash
cd backend && go get github.com/testcontainers/testcontainers-go
cd backend && go get github.com/testcontainers/testcontainers-go/modules/postgres
cd backend && go get github.com/testcontainers/testcontainers-go/modules/mongodb
```

- [ ] **Step 2: Create testcontainers helper**

```go
//go:build integration

package helpers

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainers holds test database connections.
type TestContainers struct {
	PgPool       *pgxpool.Pool
	MongoDB      *mongo.Database
	pgContainer  *tcpostgres.PostgresContainer
	mongoContainer *tcmongodb.MongoDBContainer
}

// SetupContainers starts Postgres + MongoDB containers and runs migrations.
func SetupContainers() (*TestContainers, error) {
	ctx := context.Background()
	tc := &TestContainers{}

	// Start Postgres
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("guimba_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("starting postgres container: %w", err)
	}
	tc.pgContainer = pgContainer

	pgDSN, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("getting postgres DSN: %w", err)
	}

	// Run migrations
	_, currentFile, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "migrations")
	migrationsPath = filepath.ToSlash(migrationsPath)

	m, err := migrate.New("file://"+migrationsPath, pgDSN)
	if err != nil {
		return nil, fmt.Errorf("creating migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return nil, srcErr
	}
	if dbErr != nil {
		return nil, dbErr
	}

	// Connect pgx pool
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	tc.PgPool = pool

	// Start MongoDB
	mongoContainer, err := tcmongodb.Run(ctx, "mongo:7")
	if err != nil {
		return nil, fmt.Errorf("starting mongodb container: %w", err)
	}
	tc.mongoContainer = mongoContainer

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting mongo URI: %w", err)
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("connecting to mongo: %w", err)
	}
	tc.MongoDB = mongoClient.Database("guimba_test")

	return tc, nil
}

// TruncateAll clears all tables for test isolation.
func (tc *TestContainers) TruncateAll(ctx context.Context) error {
	tables := []string{
		"beneficiary_programs",
		"transactions",
		"debts",
		"properties",
		"tenants",
		"user_roles",
		"beneficiaries",
		"users",
		"programs",
	}
	for _, table := range tables {
		if _, err := tc.PgPool.Exec(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("truncating %s: %w", table, err)
		}
	}

	// Clear MongoDB collections
	if err := tc.MongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		log.Printf("warning: could not drop audit_logs: %v", err)
	}

	return nil
}

// Cleanup stops all containers.
func (tc *TestContainers) Cleanup() {
	ctx := context.Background()
	if tc.PgPool != nil {
		tc.PgPool.Close()
	}
	if tc.pgContainer != nil {
		_ = tc.pgContainer.Terminate(ctx)
	}
	if tc.mongoContainer != nil {
		_ = tc.mongoContainer.Terminate(ctx)
	}
}
```

- [ ] **Step 3: Run `go mod tidy`**

```bash
cd backend && go mod tidy
```

- [ ] **Step 4: Commit**

```bash
cd backend && git add tests/helpers/testcontainers.go go.mod go.sum
git commit -m "test(integration): add testcontainers helper for Postgres and MongoDB"
```

---

### Task 2: TestMain + Program Repo Tests

**Files:**
- Create: `backend/tests/integration/integration_test.go`
- Create: `backend/tests/integration/program_repo_test.go`

- [ ] **Step 1: Create TestMain**

```go
//go:build integration

package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/ISubamariner/guimba-go/backend/tests/helpers"
)

var (
	testPgPool  *pgxpool.Pool
	testMongoDB *mongo.Database
	testTC      *helpers.TestContainers
)

func TestMain(m *testing.M) {
	var err error
	testTC, err = helpers.SetupContainers()
	if err != nil {
		log.Fatalf("failed to setup test containers: %v", err)
	}

	testPgPool = testTC.PgPool
	testMongoDB = testTC.MongoDB

	code := m.Run()

	testTC.Cleanup()
	os.Exit(code)
}

func truncateAll(t *testing.T) {
	t.Helper()
	if err := testTC.TruncateAll(context.Background()); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
}
```

- [ ] **Step 2: Create program repo integration tests**

```go
//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
)

func newTestProgram(name string) *entity.Program {
	now := time.Now().UTC()
	status := entity.ProgramStatusActive
	return &entity.Program{
		ID:          uuid.New(),
		Name:        name,
		Description: "Test program description",
		Status:      status,
		StartDate:   now,
		EndDate:     nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestProgramRepo_Create(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	program := newTestProgram("Integration Test Program")
	err := repo.Create(ctx, program)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify it was persisted
	found, err := repo.GetByID(ctx, program.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected program, got nil")
	}
	if found.Name != program.Name {
		t.Errorf("expected name %q, got %q", program.Name, found.Name)
	}
}

func TestProgramRepo_GetByID_NotFound(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for non-existent program")
	}
}

func TestProgramRepo_List(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	// Create 3 programs
	for i := 0; i < 3; i++ {
		if err := repo.Create(ctx, newTestProgram("Program "+string(rune('A'+i)))); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	programs, total, err := repo.List(ctx, repository.ProgramFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(programs) != 3 {
		t.Errorf("expected 3 programs, got %d", len(programs))
	}
}

func TestProgramRepo_List_WithPagination(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := repo.Create(ctx, newTestProgram("Page Program "+string(rune('A'+i)))); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	programs, total, err := repo.List(ctx, repository.ProgramFilter{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(programs) != 2 {
		t.Errorf("expected 2 programs (page 1), got %d", len(programs))
	}
}

func TestProgramRepo_List_WithStatusFilter(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	active := newTestProgram("Active Program")
	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	inactive := newTestProgram("Inactive Program")
	inactiveStatus := entity.ProgramStatusInactive
	inactive.Status = inactiveStatus
	if err := repo.Create(ctx, inactive); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	activeStatus := entity.ProgramStatusActive
	programs, total, err := repo.List(ctx, repository.ProgramFilter{
		Status: &activeStatus,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 active program, got %d", total)
	}
	if len(programs) != 1 {
		t.Errorf("expected 1 program, got %d", len(programs))
	}
}

func TestProgramRepo_Update(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	program := newTestProgram("Original Name")
	if err := repo.Create(ctx, program); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	program.Name = "Updated Name"
	if err := repo.Update(ctx, program); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.GetByID(ctx, program.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.Name != "Updated Name" {
		t.Errorf("expected name %q, got %q", "Updated Name", found.Name)
	}
}

func TestProgramRepo_Delete(t *testing.T) {
	truncateAll(t)
	repo := pg.NewProgramRepoPG(testPgPool)
	ctx := context.Background()

	program := newTestProgram("To Delete")
	if err := repo.Create(ctx, program); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(ctx, program.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should return nil after soft delete
	found, err := repo.GetByID(ctx, program.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil after delete, got program")
	}
}
```

- [ ] **Step 3: Run integration tests**

```bash
cd backend && go test -tags=integration ./tests/integration/... -v -count=1
```

Expected: All program repo tests pass.

- [ ] **Step 4: Commit**

```bash
cd backend && git add tests/integration/
git commit -m "test(integration): add TestMain and program repo integration tests"
```

---

### Task 3: User + Tenant + Property Repo Tests

**Files:**
- Create: `backend/tests/integration/user_repo_test.go`
- Create: `backend/tests/integration/tenant_repo_test.go`
- Create: `backend/tests/integration/property_repo_test.go`

- [ ] **Step 1: Create user repo tests**

```go
//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
)

func newTestUser(email string) *entity.User {
	now := time.Now().UTC()
	return &entity.User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     "Test User",
		PasswordHash: "$2a$10$dummyhashforintegrationtesting000000000000000000",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestUserRepo_Create_And_GetByEmail(t *testing.T) {
	truncateAll(t)
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	user := newTestUser("test@example.com")
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Email != "test@example.com" {
		t.Errorf("expected email %q, got %q", "test@example.com", found.Email)
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	truncateAll(t)
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	user := newTestUser("getbyid@example.com")
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.ID != user.ID {
		t.Errorf("expected ID %s, got %s", user.ID, found.ID)
	}
}

func TestUserRepo_List(t *testing.T) {
	truncateAll(t)
	repo := pg.NewUserRepoPG(testPgPool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		email := "list" + string(rune('a'+i)) + "@example.com"
		if err := repo.Create(ctx, newTestUser(email)); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	users, total, err := repo.List(ctx, repository.UserFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}
```

- [ ] **Step 2: Create tenant repo tests**

```go
//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
)

func newTestTenant(name string, landlordID uuid.UUID) *entity.Tenant {
	now := time.Now().UTC()
	return &entity.Tenant{
		ID:         uuid.New(),
		LandlordID: landlordID,
		FullName:   name,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func createTestLandlord(t *testing.T) uuid.UUID {
	t.Helper()
	repo := pg.NewUserRepoPG(testPgPool)
	user := newTestUser("landlord-" + uuid.New().String()[:8] + "@example.com")
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("creating landlord: %v", err)
	}
	return user.ID
}

func TestTenantRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	landlordID := createTestLandlord(t)
	repo := pg.NewTenantRepoPG(testPgPool)

	tenant := newTestTenant("Test Tenant", landlordID)
	if err := repo.Create(ctx, tenant); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected tenant, got nil")
	}
	if found.FullName != "Test Tenant" {
		t.Errorf("expected name %q, got %q", "Test Tenant", found.FullName)
	}
}

func TestTenantRepo_List_ScopedByLandlord(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	repo := pg.NewTenantRepoPG(testPgPool)

	landlord1 := createTestLandlord(t)
	landlord2 := createTestLandlord(t)

	if err := repo.Create(ctx, newTestTenant("Tenant A", landlord1)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(ctx, newTestTenant("Tenant B", landlord1)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(ctx, newTestTenant("Tenant C", landlord2)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	tenants, total, err := repo.List(ctx, repository.TenantFilter{
		LandlordID: &landlord1,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 tenants for landlord1, got %d", total)
	}
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(tenants))
	}
}
```

- [ ] **Step 3: Create property repo tests**

```go
//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
)

func newTestProperty(name string, landlordID uuid.UUID) *entity.Property {
	now := time.Now().UTC()
	return &entity.Property{
		ID:           uuid.New(),
		LandlordID:   landlordID,
		Name:         name,
		PropertyCode: "P-" + uuid.New().String()[:8],
		PropertyType: "RESIDENTIAL",
		SizeInSqm:    100.0,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestPropertyRepo_Create_And_GetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	landlordID := createTestLandlord(t)
	repo := pg.NewPropertyRepoPG(testPgPool)

	prop := newTestProperty("Test Property", landlordID)
	if err := repo.Create(ctx, prop); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected property, got nil")
	}
	if found.Name != "Test Property" {
		t.Errorf("expected name %q, got %q", "Test Property", found.Name)
	}
}

func TestPropertyRepo_List_ScopedByLandlord(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	repo := pg.NewPropertyRepoPG(testPgPool)

	landlord1 := createTestLandlord(t)
	landlord2 := createTestLandlord(t)

	if err := repo.Create(ctx, newTestProperty("Prop A", landlord1)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(ctx, newTestProperty("Prop B", landlord2)); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	props, total, err := repo.List(ctx, repository.PropertyFilter{
		LandlordID: &landlord1,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 property for landlord1, got %d", total)
	}
	if len(props) != 1 {
		t.Errorf("expected 1 property, got %d", len(props))
	}
}
```

- [ ] **Step 4: Run integration tests**

```bash
cd backend && go test -tags=integration ./tests/integration/... -v -count=1
```

Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
cd backend && git add tests/integration/
git commit -m "test(integration): add user, tenant, and property repo tests"
```

---

### Task 4: Debt + Transaction Repo Tests

**Files:**
- Create: `backend/tests/integration/debt_repo_test.go`
- Create: `backend/tests/integration/transaction_repo_test.go`

- [ ] **Step 1: Create debt repo tests**

The debt repo is the most complex — it tests the state machine (PENDING → PAID, PENDING → CANCELLED, PENDING → PARTIAL → PAID).

Read the following files before writing tests to understand the exact entity fields, repository interface, and PG implementation:
- `backend/internal/domain/entity/debt.go` — Debt entity with Money value object, DebtStatus, DebtType
- `backend/internal/domain/repository/debt_repository.go` — DebtRepository interface with DebtFilter
- `backend/internal/infrastructure/persistence/pg/debt_repo_pg.go` — PG implementation
- `backend/internal/domain/entity/money.go` — Money value object (Amount string, Currency string)

Write tests that cover:
1. Create debt and GetByID
2. List debts with tenant filter
3. Update debt status (simulate payment reducing balance)
4. Delete (soft delete)

Use the actual entity types and repository filter types from the domain layer. Create helper tenants via `createTestLandlord()` + tenant repo for FK constraints.

- [ ] **Step 2: Create transaction repo tests**

Read the following files first:
- `backend/internal/domain/entity/transaction.go` — Transaction entity
- `backend/internal/domain/repository/transaction_repository.go` — TransactionRepository interface
- `backend/internal/infrastructure/persistence/pg/transaction_repo_pg.go` — PG implementation

Write tests that cover:
1. Create transaction and GetByID
2. List transactions (by landlord or tenant)
3. Verify transaction (update is_verified flag)

- [ ] **Step 3: Run integration tests**

```bash
cd backend && go test -tags=integration ./tests/integration/... -v -count=1
```

- [ ] **Step 4: Commit**

```bash
cd backend && git add tests/integration/
git commit -m "test(integration): add debt and transaction repo tests"
```

---

### Task 5: Beneficiary + Audit Repo Tests

**Files:**
- Create: `backend/tests/integration/beneficiary_repo_test.go`
- Create: `backend/tests/integration/audit_repo_test.go`

- [ ] **Step 1: Create beneficiary repo tests**

Read first:
- `backend/internal/domain/entity/beneficiary.go`
- `backend/internal/domain/repository/beneficiary_repository.go`
- `backend/internal/infrastructure/persistence/pg/beneficiary_repo_pg.go`

Write tests covering:
1. Create beneficiary and GetByID
2. List beneficiaries with pagination
3. Enroll beneficiary in program
4. Remove beneficiary from program
5. GetByID returns enrolled programs

- [ ] **Step 2: Create audit repo tests**

Read first:
- `backend/internal/domain/entity/audit.go` (or wherever audit entry is defined)
- `backend/internal/domain/repository/audit_repository.go`
- `backend/internal/infrastructure/persistence/mongo/audit_repo_mongo.go`

This tests the MongoDB implementation. Write tests covering:
1. Log an audit entry (create)
2. Query audit entries (list with filters)

- [ ] **Step 3: Run integration tests**

```bash
cd backend && go test -tags=integration ./tests/integration/... -v -count=1
```

- [ ] **Step 4: Commit**

```bash
cd backend && git add tests/integration/
git commit -m "test(integration): add beneficiary and audit repo tests"
```

---

### Task 6: Auth API Integration Test

**Files:**
- Create: `backend/tests/integration/auth_api_test.go`

- [ ] **Step 1: Create auth API test**

This tests the full auth flow: register → login → refresh → me → logout. It requires wiring up a real HTTP server using the app's router.

Read these files first to understand the wiring:
- `backend/cmd/server/main.go` — how repos, usecases, handlers are wired
- `backend/internal/delivery/http/router/router.go` — how routes are registered
- `backend/internal/delivery/http/handler/auth_handler.go` — auth handler
- `backend/pkg/auth/jwt.go` — JWT manager

Write a test that:
1. Builds a real router with all dependencies wired to the test containers
2. Uses `httptest.NewServer` to serve it
3. Exercises: POST /auth/register → POST /auth/login → POST /auth/refresh → GET /auth/me → POST /auth/logout
4. Asserts proper status codes and response bodies
5. After logout, verifies token is rejected (401)

NOTE: This is a complex integration task. Read the wiring code in main.go carefully. You'll need to create a config, JWT manager, Redis client (or mock), and wire all deps. If the wiring is too complex, simplify by testing the handler directly with `httptest.NewRecorder` and manually constructed dependencies.

- [ ] **Step 2: Run integration tests**

```bash
cd backend && go test -tags=integration ./tests/integration/... -v -count=1
```

- [ ] **Step 3: Commit**

```bash
cd backend && git add tests/integration/
git commit -m "test(integration): add auth API integration test"
```

---

### Task 7: Verify Full Suite + Update MASTERPLAN

**Files:**
- Modify: `MASTERPLAN.md`

- [ ] **Step 1: Run the complete integration test suite**

```bash
cd backend && go test -tags=integration ./tests/integration/... -v -count=1
```

Expected: All integration tests pass.

- [ ] **Step 2: Also verify unit tests still pass**

```bash
cd backend && go test ./tests/unit/... -count=1
```

Expected: All 254 unit tests still pass.

- [ ] **Step 3: Update MASTERPLAN.md**

Mark integration tests as complete in Phase 4:

Change:
```
- [ ] Write integration tests in `tests/integration/`
```
To:
```
- [x] Write integration tests in `tests/integration/` — testcontainers-go (Postgres + MongoDB), covers all 9 repo implementations + auth API flow
```

- [ ] **Step 4: Commit**

```bash
git add MASTERPLAN.md
git commit -m "docs: mark integration tests complete in MASTERPLAN"
```
