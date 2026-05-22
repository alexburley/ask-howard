//go:build functional

package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	s3adapter "github.com/alexburley/ask-howard/internal/adapter/outbound/s3"
	"github.com/alexburley/ask-howard/internal/service"
	"github.com/alexburley/ask-howard/internal/testutil"
	"github.com/stretchr/testify/suite"
)

type DocumentSuite struct {
	testutil.Suite
	server *httptest.Server
}

func (s *DocumentSuite) SetupSuite() {
	s.Suite.SetupSuite()

	endpoint := os.Getenv("TEST_S3_ENDPOINT")
	if endpoint == "" {
		s.T().Skip("TEST_S3_ENDPOINT not set — skipping document functional tests")
	}

	store, err := s3adapter.NewStore(s.T().Context(), &s3adapter.Config{
		Endpoint:     endpoint,
		Bucket:       os.Getenv("TEST_S3_BUCKET"),
		Region:       os.Getenv("TEST_S3_REGION"),
		AccessKey:    os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:    os.Getenv("TEST_S3_SECRET_KEY"),
		UsePathStyle: true,
	})
	s.Require().NoError(err)

	authSvc := service.NewAuthService(postgres.NewUserRepository(s.Pool))
	docSvc := service.NewDocumentService(postgres.NewDocumentRepository(s.Pool), store)

	srv := httpserver.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), s.Pool, authSvc, docSvc, testJWTSecret)
	s.server = httptest.NewServer(srv)
}

func (s *DocumentSuite) TearDownSuite() {
	s.server.Close()
	s.Suite.TearDownSuite()
}

func TestDocumentSuite(t *testing.T) {
	suite.Run(t, new(DocumentSuite))
}

func (s *DocumentSuite) TestUploadSlot_CreatedWithPresignedURL() {
	token := s.registerAndGetToken("doc-upload@example.com")

	resp := s.postUpload(token, "ancestry.zip")
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.NotEmpty(body["document_set_id"])
	s.NotEmpty(body["presigned_url"])
	s.NotEmpty(body["object_key"])
}

func (s *DocumentSuite) TestUploadSlot_UnauthorizedWithNoCookie() {
	resp := s.postUploadNoAuth("ancestry.zip")
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *DocumentSuite) TestCompleteUpload_TransitionsToProcessing() {
	token := s.registerAndGetToken("doc-complete@example.com")

	uploadResp := s.postUpload(token, "docs.zip")
	defer uploadResp.Body.Close()
	s.Require().Equal(http.StatusCreated, uploadResp.StatusCode)

	var slot map[string]string
	s.Require().NoError(json.NewDecoder(uploadResp.Body).Decode(&slot))

	resp := s.postComplete(token, slot["document_set_id"])
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("PROCESSING", body["status"])
}

func (s *DocumentSuite) TestCompleteUpload_NotFoundForAnotherUser() {
	ownerToken := s.registerAndGetToken("doc-owner@example.com")
	otherToken := s.registerAndGetToken("doc-other@example.com")

	uploadResp := s.postUpload(ownerToken, "secret.zip")
	defer uploadResp.Body.Close()
	s.Require().Equal(http.StatusCreated, uploadResp.StatusCode)

	var slot map[string]string
	s.Require().NoError(json.NewDecoder(uploadResp.Body).Decode(&slot))

	resp := s.postComplete(otherToken, slot["document_set_id"])
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *DocumentSuite) TestGetDocumentSet_ReturnsStatus() {
	token := s.registerAndGetToken("doc-get@example.com")

	uploadResp := s.postUpload(token, "family.zip")
	defer uploadResp.Body.Close()
	s.Require().Equal(http.StatusCreated, uploadResp.StatusCode)

	var slot map[string]string
	s.Require().NoError(json.NewDecoder(uploadResp.Body).Decode(&slot))

	resp := s.getDocumentSet(token, slot["document_set_id"])
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("UPLOADING", body["status"])
	s.Equal("family.zip", body["original_filename"])
}

func (s *DocumentSuite) TestGetDocumentSet_NotFoundForAnotherUser() {
	ownerToken := s.registerAndGetToken("doc-get-owner@example.com")
	otherToken := s.registerAndGetToken("doc-get-other@example.com")

	uploadResp := s.postUpload(ownerToken, "private.zip")
	defer uploadResp.Body.Close()
	s.Require().Equal(http.StatusCreated, uploadResp.StatusCode)

	var slot map[string]string
	s.Require().NoError(json.NewDecoder(uploadResp.Body).Decode(&slot))

	resp := s.getDocumentSet(otherToken, slot["document_set_id"])
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *DocumentSuite) registerAndGetToken(email string) *http.Cookie {
	s.T().Helper()
	resp := postRegister(s.T(), s.server, email, "password123")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	c := cookieByName(resp, "token")
	s.Require().NotNil(c)
	return c
}

func (s *DocumentSuite) postUpload(token *http.Cookie, filename string) *http.Response {
	s.T().Helper()
	body, _ := json.Marshal(map[string]string{"filename": filename})
	req, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/documents/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(token)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *DocumentSuite) postUploadNoAuth(filename string) *http.Response {
	s.T().Helper()
	body, _ := json.Marshal(map[string]string{"filename": filename})
	resp, err := s.server.Client().Post(s.server.URL+"/api/documents/upload", "application/json", bytes.NewReader(body))
	s.Require().NoError(err)
	return resp
}

func (s *DocumentSuite) postComplete(token *http.Cookie, setID string) *http.Response {
	s.T().Helper()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/documents/sets/%s/complete", s.server.URL, setID), nil)
	req.AddCookie(token)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *DocumentSuite) getDocumentSet(token *http.Cookie, setID string) *http.Response {
	s.T().Helper()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/documents/sets/%s", s.server.URL, setID), nil)
	req.AddCookie(token)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	return resp
}

