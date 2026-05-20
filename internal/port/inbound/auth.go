package inbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (domain.User, error)
	Login(ctx context.Context, email, password string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}
