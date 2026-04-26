package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, input *model.LoginInput) (*sqlc.User, error)
}

type ServiceImpl struct {
	queries     *sqlc.Queries
	userService user.Service
}

func NewService(queries *sqlc.Queries, us user.Service) Service {
	return &ServiceImpl{
		queries:     queries,
		userService: us,
	}
}

func (s *ServiceImpl) Login(ctx context.Context, input *model.LoginInput) (*sqlc.User, error) {
	if strings.TrimSpace(input.Email) == "" {
		return nil, &apperror.ValidationError{Field: "email", Message: "email is required"}
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, &apperror.ValidationError{Field: "password", Message: "password is required"}
	}

	if len(input.Password) < 8 {
		return nil, &apperror.ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	if len(input.Password) > 72 {
		return nil, &apperror.ValidationError{Field: "password", Message: "password must be at most 72 characters"}
	}

	user, err := s.userService.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("failed to compare hashed passwords: %w", err)
	}

	return user, nil
}
