package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres/db"
	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type UserRepository struct {
	queries *db.Queries
}

var _ outbound.UserRepository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: db.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, params outbound.CreateUserParams) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        params.Email.String(),
		PasswordHash: params.PasswordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(&row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email.String())
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return toDomainUser(&row)
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.queries.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return toDomainUser(&row)
}

func (r *UserRepository) FindCredentialsByEmail(ctx context.Context, email domain.Email) (outbound.UserCredentials, error) {
	row, err := r.queries.FindUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return outbound.UserCredentials{}, domain.ErrInvalidCredentials
		}
		return outbound.UserCredentials{}, fmt.Errorf("find user credentials: %w", err)
	}
	user, err := toDomainUser(&row)
	if err != nil {
		return outbound.UserCredentials{}, err
	}
	return outbound.UserCredentials{User: user, PasswordHash: row.PasswordHash}, nil
}

func toDomainUser(u *db.User) (domain.User, error) {
	email, err := domain.NewEmail(u.Email)
	if err != nil {
		return domain.User{}, fmt.Errorf("parse stored email: %w", err)
	}
	return domain.User{ID: u.ID, Email: email}, nil
}
