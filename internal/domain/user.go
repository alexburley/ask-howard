package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type User struct {
	ID    uuid.UUID
	Email Email
}
