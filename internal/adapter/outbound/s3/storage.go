package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/alexburley/pulse/internal/port/outbound"
)

type ObjectStore struct {
	endpoint string
	bucket   string
}

var _ outbound.ObjectStore = (*ObjectStore)(nil)

func NewObjectStore(endpoint, bucket string) *ObjectStore {
	return &ObjectStore{endpoint: endpoint, bucket: bucket}
}

func (s *ObjectStore) Put(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error {
	panic("not implemented")
}

func (s *ObjectStore) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	panic("not implemented")
}

func (s *ObjectStore) Delete(_ context.Context, _ string) error {
	panic("not implemented")
}

func (s *ObjectStore) URL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
}
