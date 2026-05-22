package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/google/uuid"
)

type DocumentService struct {
	docs  outbound.DocumentRepository
	store outbound.ObjectStore
}

var _ inbound.DocumentService = (*DocumentService)(nil)

func NewDocumentService(docs outbound.DocumentRepository, store outbound.ObjectStore) *DocumentService {
	return &DocumentService{docs: docs, store: store}
}

func (s *DocumentService) CreateUploadSlot(ctx context.Context, userID uuid.UUID, filename string) (inbound.UploadSlotResult, error) {
	key := fmt.Sprintf("sets/%s/%s.zip", userID, uuid.New())

	presignURL, err := s.store.PresignPut(ctx, key, "application/zip", 15*time.Minute)
	if err != nil {
		return inbound.UploadSlotResult{}, fmt.Errorf("presign put: %w", err)
	}

	set, err := s.docs.CreateDocumentSet(ctx, outbound.CreateDocumentSetParams{
		UserID:           userID,
		OriginalFilename: filename,
		Status:           domain.DocumentSetStatusUploading,
		ObjectKey:        key,
	})
	if err != nil {
		return inbound.UploadSlotResult{}, fmt.Errorf("create document set: %w", err)
	}

	return inbound.UploadSlotResult{
		DocumentSetID: set.ID,
		PresignedURL:  presignURL,
		ObjectKey:     key,
	}, nil
}

func (s *DocumentService) CompleteUpload(ctx context.Context, setID, userID uuid.UUID) (domain.DocumentSet, error) {
	_, err := s.docs.GetDocumentSetByIDAndUser(ctx, setID, userID)
	if err != nil {
		return domain.DocumentSet{}, fmt.Errorf("get document set: %w", err)
	}

	set, err := s.docs.UpdateDocumentSetStatus(ctx, setID, domain.DocumentSetStatusProcessing, "")
	if err != nil {
		return domain.DocumentSet{}, fmt.Errorf("update document set status: %w", err)
	}

	return set, nil
}

func (s *DocumentService) GetDocumentSet(ctx context.Context, setID, userID uuid.UUID) (domain.DocumentSet, error) {
	set, err := s.docs.GetDocumentSetByIDAndUser(ctx, setID, userID)
	if err != nil {
		return domain.DocumentSet{}, fmt.Errorf("get document set: %w", err)
	}
	return set, nil
}
