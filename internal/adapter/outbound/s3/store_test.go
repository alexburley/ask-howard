//go:build functional

package s3_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexburley/ask-howard/internal/adapter/outbound/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *s3.Store {
	t.Helper()
	endpoint := os.Getenv("TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("TEST_S3_ENDPOINT not set — skipping S3 functional tests")
	}
	store, err := s3.NewStore(context.Background(), &s3.Config{
		Endpoint:     endpoint,
		Bucket:       os.Getenv("TEST_S3_BUCKET"),
		Region:       os.Getenv("TEST_S3_REGION"),
		AccessKey:    os.Getenv("TEST_S3_ACCESS_KEY"),
		SecretKey:    os.Getenv("TEST_S3_SECRET_KEY"),
		UsePathStyle: true,
	})
	require.NoError(t, err)
	return store
}

func TestStore_PutGetDelete(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	key := "test/" + uuid.NewString()
	want := []byte("hello ancestry")

	require.NoError(t, store.Put(ctx, key, bytes.NewReader(want), int64(len(want)), "text/plain"))

	rc, err := store.Get(ctx, key)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, rc.Close())
	require.NoError(t, err)
	require.Equal(t, want, got)

	require.NoError(t, store.Delete(ctx, key))

	_, err = store.Get(ctx, key)
	require.Error(t, err, "expected error after deletion")
}

func TestStore_PresignPutThenGet(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	key := "test/" + uuid.NewString()
	want := []byte("presigned content")

	putURL, err := store.PresignPut(ctx, key, "text/plain", time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, putURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, bytes.NewReader(want))
	require.NoError(t, err)
	req.ContentLength = int64(len(want))
	req.Header.Set("Content-Type", "text/plain")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode)

	getURL, err := store.PresignGet(ctx, key, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, getURL)

	getRes, err := http.Get(getURL) //nolint:noctx
	require.NoError(t, err)
	defer getRes.Body.Close()
	got, err := io.ReadAll(getRes.Body)
	require.NoError(t, err)
	require.Equal(t, want, got)

	require.NoError(t, store.Delete(ctx, key))
}
