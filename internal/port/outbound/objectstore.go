package outbound

import (
	"context"
	"io"
	"time"
)

// ObjectStore is the driven port for S3-compatible blob storage.
type ObjectStore interface {
	PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
}
