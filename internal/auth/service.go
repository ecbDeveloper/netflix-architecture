package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, input model.LoginInput) (sqlc.User, error)
}

type ServiceImpl struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) Service {
	return &ServiceImpl{
		queries: queries,
	}
}

func (s *ServiceImpl) Login(ctx context.Context, input model.LoginInput) (sqlc.User, error) {
	if strings.TrimSpace(input.Email) == "" {
		return sqlc.User{}, &apperror.ValidationError{Field: "email", Message: "email is required"}
	}
	if strings.TrimSpace(input.Password) == "" {
		return sqlc.User{}, &apperror.ValidationError{Field: "password", Message: "password is required"}
	}

	user, err := s.queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, &apperror.NotFoundError{Entity: "user"}
		}

		return sqlc.User{}, fmt.Errorf("failed to fetch user by email from database: %w", err)

	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return sqlc.User{}, fmt.Errorf("failed to compare hashed passwords: %w", err)
	}

	return user, nil
}
