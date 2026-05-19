package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrEmailTaken = errors.New("email already registered")

type User struct {
	ID    uuid.UUID
	Email Email
}
