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
	jobs  outbound.JobEnqueuer
}

var _ inbound.DocumentService = (*DocumentService)(nil)

func NewDocumentService(docs outbound.DocumentRepository, store outbound.ObjectStore, jobs outbound.JobEnqueuer) *DocumentService {
	return &DocumentService{docs: docs, store: store, jobs: jobs}
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

	if err := s.jobs.EnqueueExtraction(ctx, setID, userID); err != nil {
		return domain.DocumentSet{}, fmt.Errorf("enqueue extraction: %w", err)
	}

	return set, nil
}

func (s *DocumentService) GetDocumentSet(ctx context.Context, setID, userID uuid.UUID) (inbound.DocumentSetWithCount, error) {
	set, err := s.docs.GetDocumentSetByIDAndUser(ctx, setID, userID)
	if err != nil {
		return inbound.DocumentSetWithCount{}, fmt.Errorf("get document set: %w", err)
	}

	count, err := s.docs.CountDocumentsBySetID(ctx, setID)
	if err != nil {
		return inbound.DocumentSetWithCount{}, fmt.Errorf("count documents: %w", err)
	}

	return inbound.DocumentSetWithCount{DocumentSet: set, DocumentCount: count}, nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, userID uuid.UUID) ([]domain.Document, error) {
	docs, err := s.docs.ListDocumentsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	return docs, nil
}
