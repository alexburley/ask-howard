//go:build functional

package handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/service"
	"github.com/alexburley/ask-howard/internal/testutil"
	"github.com/stretchr/testify/suite"
)

type HealthSuite struct {
	testutil.Suite
	server *httptest.Server
}

func (s *HealthSuite) SetupSuite() {
	s.Suite.SetupSuite()

	authSvc := service.NewAuthService(postgres.NewUserRepository(s.Pool))
	srv := httpserver.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), s.Pool, authSvc, &noopDocumentService{}, testJWTSecret)
	s.server = httptest.NewServer(srv)
}

func (s *HealthSuite) TearDownSuite() {
	s.server.Close()
	s.Suite.TearDownSuite()
}

func TestHealthSuite(t *testing.T) {
	suite.Run(t, new(HealthSuite))
}

func (s *HealthSuite) TestHealth_ReturnsOK() {
	resp, err := s.server.Client().Get(s.server.URL + "/api/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("OK", body["status"])
}
