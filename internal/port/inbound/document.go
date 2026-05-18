package inbound

import (
	"context"

	"github.com/alexburley/pulse/internal/domain"
)

type DocumentService interface {
	Upload(ctx context.Context, title string, data []byte, contentType string) (*domain.Document, error)
	GetByID(ctx context.Context, id string) (*domain.Document, error)
	List(ctx context.Context) ([]*domain.Document, error)
}
