package domain

import "errors"

const minPasswordLength = 8

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

func (p Password) String() string { return p.value }
