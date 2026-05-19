package testutil

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// Suite is a base test suite that provides an isolated Postgres database.
// Embed it in your suite struct and call s.Suite.SetupSuite() if you override SetupSuite.
type Suite struct {
	suite.Suite
	Pool *pgxpool.Pool
}

func (s *Suite) SetupSuite() {
	connStr := NewDatabase(s.T())

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		s.T().Fatalf("create connection pool: %v", err)
	}

	s.Pool = pool
}

func (s *Suite) TearDownSuite() {
	if s.Pool != nil {
		s.Pool.Close()
	}
}
