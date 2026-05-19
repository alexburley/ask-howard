package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type UserRepository struct {
	db *pgxpool.Pool
}

var _ outbound.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash`,
		email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(_ context.Context, _ string) (domain.User, error) {
	panic("not implemented")
}

func (r *UserRepository) FindByID(_ context.Context, _ string) (domain.User, error) {
	panic("not implemented")
}
