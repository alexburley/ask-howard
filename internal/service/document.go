package service

import (
	"context"
	"fmt"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/internal/port/outbound"
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
	doc, err := s.docs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find document: %w", err)
	}
	return doc, nil
}

func (s *DocumentService) List(ctx context.Context) ([]*domain.Document, error) {
	docs, err := s.docs.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	return docs, nil
}
