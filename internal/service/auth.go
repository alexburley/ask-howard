package service

import (
	"context"
	"fmt"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users outbound.UserRepository
}

var _ inbound.AuthService = (*AuthService)(nil)

func NewAuthService(users outbound.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Login(ctx context.Context, rawEmail, rawPassword string) (domain.User, error) {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	creds, err := s.users.FindCredentialsByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("find credentials: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(rawPassword)); err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	return creds.User, nil
}

func (s *AuthService) Register(ctx context.Context, rawEmail, rawPassword string) (domain.User, error) {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return domain.User{}, fmt.Errorf("validate email: %w", err)
	}

	password, err := domain.NewPassword(rawPassword)
	if err != nil {
		return domain.User{}, fmt.Errorf("validate password: %w", err)
	}

	hash, err := password.Hash()
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, outbound.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
