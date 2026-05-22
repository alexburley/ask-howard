package outbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/google/uuid"
)

type CreateDocumentSetParams struct {
	UserID           uuid.UUID
	OriginalFilename string
	Status           string
	ObjectKey        string
}

type DocumentRepository interface {
	CreateDocumentSet(ctx context.Context, params CreateDocumentSetParams) (domain.DocumentSet, error)
	GetDocumentSetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (domain.DocumentSet, error)
	UpdateDocumentSetStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) (domain.DocumentSet, error)
}
