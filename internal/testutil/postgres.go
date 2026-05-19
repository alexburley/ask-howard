package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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
		postgres.WithDatabase("ask-howard_test"),
		postgres.WithUsername("ask-howard"),
		postgres.WithPassword("ask-howard"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
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

func applyMigrations(t *testing.T, connStr string) {
	t.Helper()

	root, err := projectRoot()
	if err != nil {
		t.Fatalf("find project root: %v", err)
	}

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

// projectRoot walks up from the working directory until it finds a go.mod file.
func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
