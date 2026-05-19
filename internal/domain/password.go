package domain

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost        = 12
	minPasswordLength = 8
)

var ErrPasswordTooShort = errors.New("password must be at least 8 characters")

type Password struct {
	value string
}

func NewPassword(s string) (Password, error) {
	if len(s) < minPasswordLength {
		return Password{}, ErrPasswordTooShort
	}
	return Password{value: s}, nil
}

func (p Password) Hash() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.value), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}
