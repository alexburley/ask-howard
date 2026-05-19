package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

var errNotImplemented = errors.New("not implemented")

type UserRepository struct {
	db *pgxpool.Pool
}

var _ outbound.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

type storedUser struct {
	id           uuid.UUID
	email        string
	passwordHash string
}

func (r *UserRepository) Create(ctx context.Context, params outbound.CreateUserParams) (domain.User, error) {
	var stored storedUser

	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash`,
		params.Email.String(), params.PasswordHash,
	).Scan(&stored.id, &stored.email, &stored.passwordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	email, err := domain.NewEmail(stored.email)
	if err != nil {
		return domain.User{}, fmt.Errorf("parse stored email: %w", err)
	}

	return domain.User{ID: stored.id, Email: email}, nil
}

func (r *UserRepository) FindByEmail(_ context.Context, _ domain.Email) (domain.User, error) {
	return domain.User{}, fmt.Errorf("find user by email: %w", errNotImplemented)
}

func (r *UserRepository) FindByID(_ context.Context, _ uuid.UUID) (domain.User, error) {
	return domain.User{}, fmt.Errorf("find user by id: %w", errNotImplemented)
}
