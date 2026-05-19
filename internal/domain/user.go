package domain

import "errors"

var ErrEmailTaken = errors.New("email already registered")

type User struct {
	ID           string
	Email        string
	PasswordHash string
}
