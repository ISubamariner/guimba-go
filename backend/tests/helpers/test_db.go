package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB provides a PostgreSQL connection pool for integration tests.
// It reads POSTGRES_DSN from the environment (set via docker-compose or CI).
type TestDB struct {
	Pool *pgxpool.Pool
}

// NewTestDB creates a test database connection.
// Falls back to default DSN if env var is not set.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://guimba:guimba_secret@localhost:5432/guimba_db?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	return &TestDB{Pool: pool}
}

// Close cleans up the test database connection.
func (db *TestDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// TruncateTable removes all rows from a table (for test cleanup).
func (db *TestDB) TruncateTable(t *testing.T, table string) {
	t.Helper()
	_, err := db.Pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	if err != nil {
		t.Fatalf("failed to truncate table %s: %v", table, err)
	}
}
