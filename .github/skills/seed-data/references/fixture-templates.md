# Seed Fixture Templates

## Example SQL Seed File (Roles & Users)

```sql
-- tests/fixtures/pg/001_roles.sql
-- Seed roles for Guimba RBAC
INSERT INTO roles (id, name, description, created_at, updated_at) VALUES
  ('role-admin',   'admin',   'Full system access',         NOW(), NOW()),
  ('role-manager', 'manager', 'Manage staff and records',   NOW(), NOW()),
  ('role-staff',   'staff',   'Create and edit records',    NOW(), NOW()),
  ('role-viewer',  'viewer',  'Read-only dashboard access', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
```

```sql
-- tests/fixtures/pg/002_users.sql
-- All seed user passwords: "testpass123"
-- Hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
INSERT INTO users (id, email, first_name, last_name, password_hash, role_id, created_at, updated_at) VALUES
  ('user-admin',   'admin@guimba.gov',   'Admin',   'User',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'role-admin',   NOW(), NOW()),
  ('user-manager', 'manager@guimba.gov', 'Manager', 'User',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'role-manager', NOW(), NOW()),
  ('user-staff',   'staff@guimba.gov',   'Staff',   'User',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'role-staff',   NOW(), NOW()),
  ('user-viewer',  'viewer@guimba.gov',  'Viewer',  'User',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'role-viewer',  NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
```

```sql
-- tests/fixtures/pg/003_borrowers.sql
INSERT INTO borrowers (id, first_name, last_name, address, contact_number, created_at, updated_at) VALUES
  ('borrower-001', 'Juan',  'Dela Cruz',   'Barangay 1, Guimba',  '09171234567', NOW(), NOW()),
  ('borrower-002', 'Maria', 'Santos',      'Barangay 2, Guimba',  '09189876543', NOW(), NOW()),
  ('borrower-003', 'Pedro', 'Reyes',       'Barangay 3, Guimba',  '09201112233', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
```

## Example JSON Seed File (Audit Logs)

```json
[
  {
    "action": "USER_LOGIN",
    "actor_id": "user-admin",
    "actor_email": "admin@guimba.gov",
    "resource_type": "auth",
    "details": {
      "ip": "127.0.0.1",
      "user_agent": "seed-script"
    },
    "created_at": "2026-01-01T00:00:00Z"
  },
  {
    "action": "DEBT_CREATED",
    "actor_id": "user-staff",
    "actor_email": "staff@guimba.gov",
    "resource_type": "debt",
    "resource_id": "debt-001",
    "details": {
      "amount": 5000,
      "borrower_id": "borrower-001"
    },
    "created_at": "2026-01-02T10:30:00Z"
  }
]
```

## Seed Runner Go Code Template

```go
// tests/helpers/cmd/seed/main.go
package main

import (
    "context"
    "log"
    "log/slog"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/ISubamariner/guimba-go/backend/tests/helpers"
)

func main() {
    dsn := os.Getenv("POSTGRES_DSN")
    if dsn == "" {
        dsn = "postgres://guimba:guimba_secret@localhost:5432/guimba_db?sslmode=disable"
    }

    ctx := context.Background()
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        log.Fatalf("connect to DB: %v", err)
    }
    defer pool.Close()

    if err := helpers.SeedPostgres(ctx, pool, "tests/fixtures"); err != nil {
        log.Fatalf("seed failed: %v", err)
    }

    slog.Info("seeding complete")
}
```

```go
// tests/helpers/seed.go
package helpers

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
    "sort"

    "github.com/jackc/pgx/v5/pgxpool"
)

func SeedPostgres(ctx context.Context, pool *pgxpool.Pool, fixtureDir string) error {
    files, err := filepath.Glob(filepath.Join(fixtureDir, "pg", "*.sql"))
    if err != nil {
        return fmt.Errorf("glob fixtures: %w", err)
    }
    sort.Strings(files)

    for _, f := range files {
        data, err := os.ReadFile(f)
        if err != nil {
            return fmt.Errorf("read %s: %w", filepath.Base(f), err)
        }
        if _, err := pool.Exec(ctx, string(data)); err != nil {
            return fmt.Errorf("exec %s: %w", filepath.Base(f), err)
        }
        slog.Info("seeded", "file", filepath.Base(f))
    }
    return nil
}
```
