package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{
		Queries: queries,
	}
}

func (s *Service) Login(ctx context.Context, input model.LoginInput) (uuid.UUID, error) {
	if strings.TrimSpace(input.Email) == "" {
		return uuid.UUID{}, &apperror.ValidationError{Field: "email", Message: "email is required"}
	}
	if strings.TrimSpace(input.Password) == "" {
		return uuid.UUID{}, &apperror.ValidationError{Field: "password", Message: "password is required"}
	}

	user, err := s.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, &apperror.NotFoundError{Entity: "user"}
		}

		return uuid.UUID{}, fmt.Errorf("failed to fetch user by email from database: %w", err)

	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compare hashed passwords: %w", err)
	}

	return user.ID, nil
}
