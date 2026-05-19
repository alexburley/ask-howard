package inbound

import (
	"context"

	"github.com/alexburley/ask-howard/internal/domain"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (domain.User, error)
}
