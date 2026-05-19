package outbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
)

type DocumentRepository interface {
	Save(ctx context.Context, doc *domain.Document) error
	FindByID(ctx context.Context, id string) (*domain.Document, error)
	FindAll(ctx context.Context) ([]*domain.Document, error)
}
