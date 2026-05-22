package outbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/google/uuid"
)

type CreateDocumentSetParams struct {
	UserID           uuid.UUID
	OriginalFilename string
	Status           domain.DocumentSetStatus
	ObjectKey        string
}

type InsertDocumentParams struct {
	SetID       uuid.UUID
	UserID      uuid.UUID
	Filename    string
	ContentType string
	SizeBytes   int64
	ObjectKey   string
}

type DocumentRepository interface {
	CreateDocumentSet(ctx context.Context, params CreateDocumentSetParams) (domain.DocumentSet, error)
	GetDocumentSetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (domain.DocumentSet, error)
	UpdateDocumentSetStatus(ctx context.Context, id uuid.UUID, status domain.DocumentSetStatus, errorMsg string) (domain.DocumentSet, error)
	InsertDocument(ctx context.Context, params *InsertDocumentParams) (domain.Document, error)
	ListDocumentsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Document, error)
	GetDocumentByIDAndUser(ctx context.Context, id, userID uuid.UUID) (domain.Document, error)
	DeleteDocumentsBySetID(ctx context.Context, setID uuid.UUID) error
	CountDocumentsBySetID(ctx context.Context, setID uuid.UUID) (int64, error)
}
