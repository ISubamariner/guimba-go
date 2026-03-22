---
name: seed-data
description: "Manages database seed data and test fixtures for development and testing. Use when user says 'seed database', 'add test data', 'create fixtures', 'populate database', 'sample data', 'reset and seed', or when working with tests/fixtures/ or migrations/."
---

# Seed Data & Test Fixtures

Manages database seed data for development and testing environments.

## File Locations

| Type | Location | Format |
|:---|:---|:---|
| PostgreSQL seeds | `tests/fixtures/pg/` | `.sql` files |
| MongoDB seeds | `tests/fixtures/mongo/` | `.json` files |
| Seed runner | `tests/helpers/seed.go` | Go code |

## Seed File Naming

```
tests/fixtures/
├── pg/
│   ├── 001_roles.sql
│   ├── 002_users.sql
│   ├── 003_borrowers.sql
│   └── 004_debts.sql
├── mongo/
│   ├── audit_logs.json
│   └── documents.json
└── README.md
```

Files are numbered to ensure execution order respects foreign key dependencies.

## Idempotent Seeds

Always use `ON CONFLICT DO NOTHING` or upsert patterns:

```sql
-- tests/fixtures/pg/001_roles.sql
INSERT INTO roles (id, name, description) VALUES
  ('role-admin', 'admin', 'Full system access'),
  ('role-manager', 'manager', 'Manage staff and records'),
  ('role-staff', 'staff', 'Create and edit records'),
  ('role-viewer', 'viewer', 'Read-only access')
ON CONFLICT (id) DO NOTHING;
```

## Password Hashing in Seeds

**Never store plaintext passwords** — always use pre-hashed bcrypt values:

```sql
-- tests/fixtures/pg/002_users.sql
-- Password for all seed users: "testpass123"
-- bcrypt hash generated with cost 10
INSERT INTO users (id, email, first_name, last_name, password_hash, role_id) VALUES
  ('user-admin', 'admin@guimba.gov', 'Admin', 'User',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'role-admin'),
  ('user-staff', 'staff@guimba.gov', 'Staff', 'User',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'role-staff')
ON CONFLICT (id) DO NOTHING;
```

## Seed Runner

```go
// tests/helpers/seed.go
func SeedPostgres(ctx context.Context, db *pgxpool.Pool, fixtureDir string) error {
    files, _ := filepath.Glob(filepath.Join(fixtureDir, "pg", "*.sql"))
    sort.Strings(files) // ensures order by number prefix

    for _, f := range files {
        data, err := os.ReadFile(f)
        if err != nil {
            return fmt.Errorf("read %s: %w", f, err)
        }
        if _, err := db.Exec(ctx, string(data)); err != nil {
            return fmt.Errorf("exec %s: %w", f, err)
        }
        slog.Info("seeded", "file", filepath.Base(f))
    }
    return nil
}
```

## MongoDB Seeds

```json
// tests/fixtures/mongo/audit_logs.json
[
  {
    "action": "USER_LOGIN",
    "actor_id": "user-admin",
    "actor_email": "admin@guimba.gov",
    "details": { "ip": "127.0.0.1" },
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

## Integration with Docker Compose

Seed on startup via an init script or after `docker compose up`:

```bash
# Seed after services are healthy
docker compose up -d
go run tests/helpers/cmd/seed/main.go
```

## Integration with Tests

```go
func TestMain(m *testing.M) {
    // Start testcontainers
    db := setupTestDB()
    defer db.Close()

    // Seed
    seed.SeedPostgres(context.Background(), db, "../../tests/fixtures")

    // Run tests
    code := m.Run()

    // Cleanup
    cleanupTestDB(db)
    os.Exit(code)
}
```

## Rules

- Seeds must be idempotent (safe to run multiple times)
- Never include real personal data in seeds
- Always use pre-hashed passwords (never plaintext)
- Number files to control execution order
- Keep dev seeds and test seeds separate if they differ
