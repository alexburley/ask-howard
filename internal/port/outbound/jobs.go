package outbound

import (
	"context"

	"github.com/google/uuid"
)

type JobEnqueuer interface {
	EnqueueExtraction(ctx context.Context, setID, userID uuid.UUID) error
}
