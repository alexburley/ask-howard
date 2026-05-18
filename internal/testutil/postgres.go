package testutil

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// PostgresContainer holds a running test Postgres instance.
type PostgresContainer struct {
	ConnectionString string
	container        *postgres.PostgresContainer
}

// NewPostgresContainer starts a Postgres container and registers cleanup on t.
// The returned ConnectionString is ready to pass to pgxpool.New or sql.Open.
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

	return &PostgresContainer{ConnectionString: connStr, container: c}
}
