package outbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
}
