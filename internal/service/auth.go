package service

import (
	"context"
	"fmt"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type AuthService struct {
	users outbound.UserRepository
}

var _ inbound.AuthService = (*AuthService)(nil)

func NewAuthService(users outbound.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
