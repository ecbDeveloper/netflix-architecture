package user

import (
	"errors"
	"strings"
)

type CPF string

func NewCPF(value string) (CPF, error) {
	if len(value) != 11 {
		return "", errors.New("cpf must have exactly 11 characters")
	}
	return CPF(value), nil
}

func (c CPF) String() string { return string(c) }

type Email string

func NewEmail(value string) (Email, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("email is required")
	}
	return Email(value), nil
}

func (e Email) String() string { return string(e) }

type RawPassword string

func NewRawPassword(value string) (RawPassword, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("password is required")
	}
	if len(value) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	if len(value) > 72 {
		return "", errors.New("password must be at most 72 characters")
	}
	return RawPassword(value), nil
}

func (p RawPassword) String() string { return string(p) }
