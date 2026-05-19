package testutil

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// PostgresContainer holds a running test Postgres instance with migrations applied.
type PostgresContainer struct {
	ConnectionString string
	container        *postgres.PostgresContainer
}

// NewPostgresContainer starts a Postgres container, applies all migrations, and
// registers cleanup on t. The returned ConnectionString is ready to pass to
// pgxpool.New or sql.Open.
func NewPostgresContainer(t *testing.T) *PostgresContainer {
	t.Helper()

	ctx := context.Background()

	c, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pulse_test"),
		postgres.WithUsername("pulse"),
		postgres.WithPassword("pulse"),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Errorf("terminate postgres container: %v", err)
		}
	})

	connStr, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get postgres connection string: %v", err)
	}

	applyMigrations(t, connStr)

	return &PostgresContainer{ConnectionString: connStr, container: c}
}

// applyMigrations runs all pending Atlas migrations against the given database URL.
func applyMigrations(t *testing.T, connStr string) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve source file path for migrations")
	}
	root := filepath.Join(filepath.Dir(filename), "../..")

	client, err := atlasexec.NewClient(root, "atlas")
	if err != nil {
		t.Fatalf("create atlas client: %v", err)
	}

	_, err = client.MigrateApply(context.Background(), &atlasexec.MigrateApplyParams{
		URL:    connStr,
		DirURL: "file://migrations",
	})
	if err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}
