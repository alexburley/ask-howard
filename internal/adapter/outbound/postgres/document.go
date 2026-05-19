package postgres

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

var _ outbound.DocumentRepository = (*DocumentRepository)(nil)

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Save(_ context.Context, _ *domain.Document) error {
	panic("not implemented")
}

func (r *DocumentRepository) FindByID(_ context.Context, _ string) (*domain.Document, error) {
	panic("not implemented")
}

func (r *DocumentRepository) FindAll(_ context.Context) ([]*domain.Document, error) {
	panic("not implemented")
}
