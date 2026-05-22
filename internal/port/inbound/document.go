package inbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/google/uuid"
)

type UploadSlotResult struct {
	DocumentSetID uuid.UUID
	PresignedURL  string
	ObjectKey     string
}

type DocumentService interface {
	CreateUploadSlot(ctx context.Context, userID uuid.UUID, filename string) (UploadSlotResult, error)
	CompleteUpload(ctx context.Context, setID, userID uuid.UUID) (domain.DocumentSet, error)
	GetDocumentSet(ctx context.Context, setID, userID uuid.UUID) (domain.DocumentSet, error)
}
