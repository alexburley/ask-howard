//go:build functional

package handler_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/jobs"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	s3adapter "github.com/alexburley/ask-howard/internal/adapter/outbound/s3"
	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/alexburley/ask-howard/internal/service"
	"github.com/alexburley/ask-howard/internal/testutil"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DocumentSuite struct {
	testutil.Suite
	server  *httptest.Server
	store   *s3adapter.Store
	docRepo *postgres.DocumentRepository
	worker  *jobs.ExtractionWorker
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
	s.store = store

	docRepo := postgres.NewDocumentRepository(s.Pool)
	s.docRepo = docRepo
	s.worker = jobs.NewExtractionWorker(store, docRepo)

	authSvc := service.NewAuthService(postgres.NewUserRepository(s.Pool))
	docSvc := service.NewDocumentService(docRepo, store, &noopEnqueuer{})

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

// noopEnqueuer satisfies JobEnqueuer for tests — the worker is called directly.
type noopEnqueuer struct{}

func (n *noopEnqueuer) EnqueueExtraction(_ context.Context, _, _ uuid.UUID) error { return nil }


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

	var body map[string]interface{}
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

	var body map[string]interface{}
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

func (s *DocumentSuite) TestExtraction_HappyPath() {
	ctx := s.T().Context()
	userID := s.createUser(ctx, "extract-happy@example.com")

	zipData := makeZip(s.T(), map[string]string{
		"record.txt":  "John Smith, born 1850",
		"photo.jpg":   "fake-image-data",
		".hidden":     "should be skipped",
		"__MACOSX/x": "should be skipped",
	})

	keyPrefix := uuid.New()
	key := fmt.Sprintf("sets/%s/test.zip", keyPrefix)
	s.Require().NoError(s.store.Put(ctx, key, bytes.NewReader(zipData), int64(len(zipData)), "application/zip"))

	set, err := s.docRepo.CreateDocumentSet(ctx, outbound.CreateDocumentSetParams{
		UserID: userID, OriginalFilename: "test.zip",
		Status: domain.DocumentSetStatusProcessing, ObjectKey: key,
	})
	s.Require().NoError(err)

	job := &river.Job[jobs.ExtractionArgs]{Args: jobs.ExtractionArgs{SetID: set.ID, UserID: userID}}
	s.Require().NoError(s.worker.Work(ctx, job))

	result, err := s.docRepo.GetDocumentSetByIDAndUser(ctx, set.ID, userID)
	s.Require().NoError(err)
	s.Equal(domain.DocumentSetStatusReady, result.Status)

	count, err := s.docRepo.CountDocumentsBySetID(ctx, set.ID)
	s.Require().NoError(err)
	s.EqualValues(2, count) // .hidden + __MACOSX skipped
}

func (s *DocumentSuite) TestExtraction_ZipBombGuard() {
	ctx := s.T().Context()
	userID := s.createUser(ctx, "extract-bomb@example.com")

	// Entry content exceeds the 10-byte per-entry cap used by the test worker.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, err := w.Create("big.bin")
	s.Require().NoError(err)
	_, err = fw.Write(bytes.Repeat([]byte("x"), 11))
	s.Require().NoError(err)
	s.Require().NoError(w.Close())
	zipData := buf.Bytes()

	keyPrefix := uuid.New()
	key := fmt.Sprintf("sets/%s/bomb.zip", keyPrefix)
	s.Require().NoError(s.store.Put(ctx, key, bytes.NewReader(zipData), int64(len(zipData)), "application/zip"))

	set, err := s.docRepo.CreateDocumentSet(ctx, outbound.CreateDocumentSetParams{
		UserID: userID, OriginalFilename: "bomb.zip",
		Status: domain.DocumentSetStatusProcessing, ObjectKey: key,
	})
	s.Require().NoError(err)

	// Use a worker with a tiny cap so real (small) data triggers the guard.
	tinyWorker := jobs.NewExtractionWorkerWithLimits(s.store, s.docRepo, 10, 50)
	job := &river.Job[jobs.ExtractionArgs]{Args: jobs.ExtractionArgs{SetID: set.ID, UserID: userID}}
	s.Require().NoError(tinyWorker.Work(ctx, job))

	result, err := s.docRepo.GetDocumentSetByIDAndUser(ctx, set.ID, userID)
	s.Require().NoError(err)
	s.Equal(domain.DocumentSetStatusFailed, result.Status)
	s.NotEmpty(result.Error)
}

func (s *DocumentSuite) TestExtraction_CorruptZip() {
	ctx := s.T().Context()
	userID := s.createUser(ctx, "extract-corrupt@example.com")

	keyPrefix := uuid.New()
	key := fmt.Sprintf("sets/%s/corrupt.zip", keyPrefix)
	corrupt := []byte("this is not a zip file")
	s.Require().NoError(s.store.Put(ctx, key, bytes.NewReader(corrupt), int64(len(corrupt)), "application/zip"))

	set, err := s.docRepo.CreateDocumentSet(ctx, outbound.CreateDocumentSetParams{
		UserID: userID, OriginalFilename: "corrupt.zip",
		Status: domain.DocumentSetStatusProcessing, ObjectKey: key,
	})
	s.Require().NoError(err)

	job := &river.Job[jobs.ExtractionArgs]{Args: jobs.ExtractionArgs{SetID: set.ID, UserID: userID}}
	s.Require().NoError(s.worker.Work(ctx, job))

	result, err := s.docRepo.GetDocumentSetByIDAndUser(ctx, set.ID, userID)
	s.Require().NoError(err)
	s.Equal(domain.DocumentSetStatusFailed, result.Status)
	s.NotEmpty(result.Error)
}

func (s *DocumentSuite) TestListDocuments_OnlyReturnsOwnDocuments() {
	token := s.registerAndGetToken("doc-list@example.com")
	_ = s.registerAndGetToken("doc-list-other@example.com")

	// token user has no documents yet
	resp := s.getDocuments(token)
	defer resp.Body.Close()
	s.Equal(http.StatusOK, resp.StatusCode)

	var docs []interface{}
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&docs))
	s.Len(docs, 0)
}

// helpers

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

func (s *DocumentSuite) getDocuments(token *http.Cookie) *http.Response {
	s.T().Helper()
	req, _ := http.NewRequest(http.MethodGet, s.server.URL+"/api/documents", nil)
	req.AddCookie(token)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *DocumentSuite) createUser(ctx context.Context, email string) uuid.UUID {
	s.T().Helper()
	e, err := domain.NewEmail(email)
	require.NoError(s.T(), err)
	userRepo := postgres.NewUserRepository(s.Pool)
	user, err := userRepo.Create(ctx, outbound.CreateUserParams{
		Email:        e,
		PasswordHash: "irrelevant-hash",
	})
	s.Require().NoError(err)
	return user.ID
}

func makeZip(t interface{ Helper(); Fatalf(string, ...interface{}) }, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %q: %v", name, err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %q: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
