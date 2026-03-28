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
