package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("integration tests are scaffolded; set DATABASE_URL to run")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("build db pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestUploadRepositoryIntegration(t *testing.T) {
	t.Skip("integration tests are scaffolded; add migrations/fixtures before enabling")
	_ = testDBPool(t)
}

func TestFileRepositoryIntegration(t *testing.T) {
	t.Skip("integration tests are scaffolded; add migrations/fixtures before enabling")
	_ = testDBPool(t)
}

func TestJobRepositoryIntegration(t *testing.T) {
	t.Skip("integration tests are scaffolded; add migrations/fixtures before enabling")
	_ = testDBPool(t)
}

func TestParsedRecordRepositoryIntegration(t *testing.T) {
	t.Skip("integration tests are scaffolded; add migrations/fixtures before enabling")
	_ = testDBPool(t)
}

func TestDashboardRepositoryIntegration(t *testing.T) {
	t.Skip("integration tests are scaffolded; add migrations/fixtures before enabling")
	_ = testDBPool(t)
}

