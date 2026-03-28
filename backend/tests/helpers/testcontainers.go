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
	_, err := tc.PgPool.Exec(ctx, `TRUNCATE TABLE
		beneficiary_programs,
		transactions,
		debts,
		properties,
		tenants,
		user_roles,
		beneficiaries,
		users,
		programs
		CASCADE`)
	if err != nil {
		return fmt.Errorf("truncating tables: %w", err)
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
	if tc.MongoDB != nil {
		if err := tc.MongoDB.Client().Disconnect(ctx); err != nil {
			log.Printf("warning: failed to disconnect mongo client: %v", err)
		}
	}
	if tc.pgContainer != nil {
		if err := tc.pgContainer.Terminate(ctx); err != nil {
			log.Printf("warning: failed to terminate postgres container: %v", err)
		}
	}
	if tc.mongoContainer != nil {
		if err := tc.mongoContainer.Terminate(ctx); err != nil {
			log.Printf("warning: failed to terminate mongo container: %v", err)
		}
	}
}
