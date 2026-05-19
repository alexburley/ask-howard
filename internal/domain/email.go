package domain

import (
	"errors"
	"net/mail"
)

var ErrInvalidEmail = errors.New("invalid email address")

type Email struct {
	value string
}

func NewEmail(s string) (Email, error) {
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: s}, nil
}

func (e Email) String() string { return e.value }
