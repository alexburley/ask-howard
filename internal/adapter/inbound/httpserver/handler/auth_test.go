//go:build functional

package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/service"
	"github.com/alexburley/ask-howard/internal/testutil"
	"github.com/stretchr/testify/suite"
)

var testJWTSecret = auth.NewJWTSecret("test-secret-at-least-32-chars-long")

type AuthSuite struct {
	testutil.Suite
	server *httptest.Server
}

func (s *AuthSuite) SetupSuite() {
	s.Suite.SetupSuite()

	authSvc := service.NewAuthService(postgres.NewUserRepository(s.Pool))
	srv := httpserver.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), s.Pool, authSvc, testJWTSecret)
	s.server = httptest.NewServer(srv)
}

func (s *AuthSuite) TearDownSuite() {
	s.server.Close()
	s.Suite.TearDownSuite()
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}

func (s *AuthSuite) TestRegister_CreatedWithValidCredentials() {
	resp := postRegister(s.T(), s.server, "alice@example.com", "password123")
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("alice@example.com", body["email"])
	s.NotEmpty(body["id"])

	cookie := cookieByName(resp, "token")
	s.Require().NotNil(cookie, "token cookie not set")
	s.True(cookie.HttpOnly, "token cookie is not HttpOnly")
}

func (s *AuthSuite) TestRegister_ConflictOnDuplicateEmail() {
	first := postRegister(s.T(), s.server, "bob@example.com", "password123")
	first.Body.Close()
	s.Require().Equal(http.StatusCreated, first.StatusCode, "first registration should succeed")

	resp := postRegister(s.T(), s.server, "bob@example.com", "password123")
	defer resp.Body.Close()

	s.Equal(http.StatusConflict, resp.StatusCode)
}

func (s *AuthSuite) TestRegister_UnprocessableOnShortPassword() {
	resp := postRegister(s.T(), s.server, "carol@example.com", "short")
	defer resp.Body.Close()

	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func (s *AuthSuite) TestRegister_UnprocessableOnInvalidEmail() {
	resp := postRegister(s.T(), s.server, "not-an-email", "password123")
	defer resp.Body.Close()

	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
}

func postRegister(t *testing.T, ts *httptest.Server, email, password string) *http.Response {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := ts.Client().Post(ts.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/auth/register: %v", err)
	}
	return resp
}

func cookieByName(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
