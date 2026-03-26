package apperror

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

type ConflictError struct {
	Field string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s is already in use", e.Field)
}

type NotFoundError struct {
	Entity string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Entity)
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func UniqueViolationField(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return "field"
	}

	detail := pgErr.Detail
	if start := strings.Index(detail, "("); start != -1 {
		if end := strings.Index(detail[start:], ")"); end != -1 {
			return detail[start+1 : start+end]
		}
	}

	return pgErr.ConstraintName
}
