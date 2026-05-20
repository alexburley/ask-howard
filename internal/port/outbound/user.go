package outbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/google/uuid"
)

type CreateUserParams struct {
	Email        domain.Email
	PasswordHash string
}

type UserCredentials struct {
	User         domain.User
	PasswordHash string
}

type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	FindCredentialsByEmail(ctx context.Context, email domain.Email) (UserCredentials, error)
}
