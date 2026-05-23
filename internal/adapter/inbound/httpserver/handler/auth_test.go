//go:build functional

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/internal/service"
	"github.com/alexburley/ask-howard/internal/testutil"
	"github.com/google/uuid"
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
	srv := httpserver.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), s.Pool, authSvc, &noopDocumentService{}, testJWTSecret)
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

func (s *AuthSuite) TestLogin_OKWithValidCredentials() {
	postRegister(s.T(), s.server, "login-ok@example.com", "password123").Body.Close()

	resp := postLogin(s.T(), s.server, "login-ok@example.com", "password123")
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("login-ok@example.com", body["email"])
	s.NotEmpty(body["id"])

	cookie := cookieByName(resp, "token")
	s.Require().NotNil(cookie, "token cookie not set")
	s.True(cookie.HttpOnly)
}

func (s *AuthSuite) TestLogin_UnauthorizedOnWrongPassword() {
	postRegister(s.T(), s.server, "login-wrongpw@example.com", "password123").Body.Close()

	resp := postLogin(s.T(), s.server, "login-wrongpw@example.com", "wrongpassword")
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *AuthSuite) TestLogin_UnauthorizedOnUnknownEmail() {
	resp := postLogin(s.T(), s.server, "nobody@example.com", "password123")
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *AuthSuite) TestLogin_UnauthorizedOnInvalidEmailFormat() {
	resp := postLogin(s.T(), s.server, "not-an-email", "password123")
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *AuthSuite) TestMe_OKWithValidCookie() {
	registerResp := postRegister(s.T(), s.server, "me-user@example.com", "password123")
	defer registerResp.Body.Close()
	s.Require().Equal(http.StatusCreated, registerResp.StatusCode)

	req, _ := http.NewRequest(http.MethodGet, s.server.URL+"/api/auth/me", nil)
	req.AddCookie(cookieByName(registerResp, "token"))
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("me-user@example.com", body["email"])
	s.NotEmpty(body["id"])
}

func (s *AuthSuite) TestMe_UnauthorizedWithNoCookie() {
	req, _ := http.NewRequest(http.MethodGet, s.server.URL+"/api/auth/me", nil)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *AuthSuite) TestMe_UnauthorizedWithTamperedToken() {
	req, _ := http.NewRequest(http.MethodGet, s.server.URL+"/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "not.a.valid.jwt"})
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *AuthSuite) TestLogout_OKWithoutCookie() {
	resp := postLogout(s.T(), s.server)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *AuthSuite) TestLogout_ClearsCookieAfterLogin() {
	postRegister(s.T(), s.server, "logout-user@example.com", "password123").Body.Close()
	postLogin(s.T(), s.server, "logout-user@example.com", "password123").Body.Close()

	resp := postLogout(s.T(), s.server)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	cookie := cookieByName(resp, "token")
	s.Require().NotNil(cookie, "cleared token cookie not present in response")
	s.Equal(-1, cookie.MaxAge)
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

func postLogin(t *testing.T, ts *httptest.Server, email, password string) *http.Response {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := ts.Client().Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/auth/login: %v", err)
	}
	return resp
}

func postLogout(t *testing.T, ts *httptest.Server) *http.Response {
	t.Helper()

	resp, err := ts.Client().Post(ts.URL+"/api/auth/logout", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/auth/logout: %v", err)
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

type noopDocumentService struct{}

func (n *noopDocumentService) CreateUploadSlot(_ context.Context, _ uuid.UUID, _ string) (inbound.UploadSlotResult, error) {
	return inbound.UploadSlotResult{}, nil
}

func (n *noopDocumentService) CompleteUpload(_ context.Context, _, _ uuid.UUID) (inbound.DocumentSetWithCount, error) {
	return inbound.DocumentSetWithCount{}, nil
}

func (n *noopDocumentService) GetDocumentSet(_ context.Context, _, _ uuid.UUID) (inbound.DocumentSetWithCount, error) {
	return inbound.DocumentSetWithCount{}, nil
}

func (n *noopDocumentService) ListDocuments(_ context.Context, _ uuid.UUID) ([]inbound.DocumentWithURL, error) {
	return nil, nil
}

var _ inbound.DocumentService = (*noopDocumentService)(nil)
