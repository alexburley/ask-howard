package service

import (
	"context"

	"github.com/alexburley/pulse/internal/domain"
	"github.com/alexburley/pulse/internal/port/inbound"
	"github.com/alexburley/pulse/internal/port/outbound"
)

type DocumentService struct {
	docs    outbound.DocumentRepository
	objects outbound.ObjectStore
}

var _ inbound.DocumentService = (*DocumentService)(nil)

func NewDocumentService(docs outbound.DocumentRepository, objects outbound.ObjectStore) *DocumentService {
	return &DocumentService{docs: docs, objects: objects}
}

func (s *DocumentService) Upload(_ context.Context, _ string, _ []byte, _ string) (*domain.Document, error) {
	panic("not implemented")
}

func (s *DocumentService) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	return s.docs.FindByID(ctx, id)
}

func (s *DocumentService) List(ctx context.Context) ([]*domain.Document, error) {
	return s.docs.FindAll(ctx)
}
