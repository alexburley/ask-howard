package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/jackc/pgx/v5"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBUser     = "ask-howard"
	testDBPassword = "ask-howard"
)

var (
	sharedOnce  sync.Once
	sharedPG    *postgresContainer
	sharedPGErr error
	dbCounter   atomic.Int64
)

type postgresContainer struct {
	host string
	port string
}

func (c *postgresContainer) adminConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		testDBUser, testDBPassword, c.host, c.port)
}

func (c *postgresContainer) connStrFor(name string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDBUser, testDBPassword, c.host, c.port, name)
}

// getSharedContainer returns a connection to the test postgres instance.
// When TEST_DATABASE_URL is set (e.g. inside the CI Docker container) it
// connects directly to the compose postgres. Otherwise it starts a throwaway
// testcontainer — useful for running tests locally without the stack.
func getSharedContainer() (*postgresContainer, error) {
	sharedOnce.Do(func() {
		if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
			cfg, err := pgx.ParseConfig(url)
			if err != nil {
				sharedPGErr = fmt.Errorf("parse TEST_DATABASE_URL: %w", err)
				return
			}
			sharedPG = &postgresContainer{
				host: cfg.Host,
				port: fmt.Sprintf("%d", cfg.Port),
			}
			return
		}

		ctx := context.Background()

		c, err := postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("initial"),
			postgres.WithUsername(testDBUser),
			postgres.WithPassword(testDBPassword),
			tc.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2),
			),
		)
		if err != nil {
			sharedPGErr = fmt.Errorf("start postgres container: %w", err)
			return
		}

		host, err := c.Host(ctx)
		if err != nil {
			sharedPGErr = fmt.Errorf("get postgres host: %w", err)
			return
		}

		port, err := c.MappedPort(ctx, "5432")
		if err != nil {
			sharedPGErr = fmt.Errorf("get postgres port: %w", err)
			return
		}

		sharedPG = &postgresContainer{host: host, port: port.Port()}
	})

	return sharedPG, sharedPGErr
}

// NewDatabase creates a fresh database in the shared Postgres instance,
// applies all migrations, and registers cleanup (drop) on t.
func NewDatabase(t *testing.T) string {
	t.Helper()

	c, err := getSharedContainer()
	if err != nil {
		t.Fatalf("acquire shared postgres: %v", err)
	}

	ctx := context.Background()
	dbName := fmt.Sprintf("test_%d", dbCounter.Add(1))

	conn, err := pgx.Connect(ctx, c.adminConnStr())
	if err != nil {
		t.Fatalf("connect to admin database: %v", err)
	}

	if _, err := conn.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		if closeErr := conn.Close(ctx); closeErr != nil {
			t.Errorf("close admin connection: %v", closeErr)
		}
		t.Fatalf("create database %s: %v", dbName, err)
	}

	if err := conn.Close(ctx); err != nil {
		t.Errorf("close admin connection: %v", err)
	}

	connStr := c.connStrFor(dbName)
	applyMigrations(t, connStr)

	t.Cleanup(func() {
		dropConn, err := pgx.Connect(ctx, c.adminConnStr())
		if err != nil {
			t.Errorf("connect for database cleanup: %v", err)
			return
		}
		defer func() {
			if err := dropConn.Close(ctx); err != nil {
				t.Errorf("close cleanup connection: %v", err)
			}
		}()

		_, _ = dropConn.Exec(ctx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1", dbName)
		_, _ = dropConn.Exec(ctx, "DROP DATABASE "+dbName)
	})

	return connStr
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
		DirURL: "file://internal/adapter/outbound/postgres/migrations",
	})
	if err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

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
