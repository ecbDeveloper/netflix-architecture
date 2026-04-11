package user

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

type Service interface {
	CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	ListUsers(ctx context.Context) ([]*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type ServiceImpl struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) Service {
	return &ServiceImpl{
		queries: queries,
	}
}

func (s *ServiceImpl) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	if strings.TrimSpace(input.Email) == "" {
		return nil, &apperror.ValidationError{Field: "email", Message: "email is required"}
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, &apperror.ValidationError{Field: "name", Message: "name is required"}
	}
	if len(input.Cpf) != 11 {
		return nil, &apperror.ValidationError{Field: "cpf", Message: "cpf must have exactly 11 characters"}
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, &apperror.ValidationError{Field: "password", Message: "password is required"}
	}

	userID := uuid.New()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       userID,
		Email:    input.Email,
		Name:     input.Name,
		Cpf:      input.Cpf,
		Password: string(hashedPassword),
	})
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			field := apperror.UniqueViolationField(err)
			return nil, &apperror.ConflictError{Field: field}
		}
		return nil, fmt.Errorf("failed to insert user on database: %w", err)
	}

	return toGraphQLModel(user), nil
}

func (s *ServiceImpl) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "user"}
		}
		return nil, fmt.Errorf("failed to fetch user %v from database: %w", id, err)
	}

	return toGraphQLModel(user), nil
}

func (s *ServiceImpl) ListUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all users from database: %w", err)
	}

	modelUsers := make([]*model.User, len(users))
	for i, user := range users {
		modelUsers[i] = toGraphQLModel(user)
	}
	return modelUsers, nil
}

func (s *ServiceImpl) UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	if input.Email != nil && strings.TrimSpace(*input.Email) == "" {
		return nil, &apperror.ValidationError{Field: "email", Message: "email cannot be empty"}
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return nil, &apperror.ValidationError{Field: "name", Message: "name cannot be empty"}
	}
	if input.Password != nil && strings.TrimSpace(*input.Password) == "" {
		return nil, &apperror.ValidationError{Field: "password", Message: "password cannot be empty"}
	}

	updateParams := sqlc.UpdateUserParams{
		ID: id,
	}

	if input.Email != nil {
		updateParams.Email = *input.Email
	}
	if input.Name != nil {
		updateParams.Name = *input.Name
	}
	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		updateParams.Password = string(hashedPassword)
	}

	user, err := s.queries.UpdateUser(ctx, updateParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "user"}
		}
		if apperror.IsUniqueViolation(err) {
			field := apperror.UniqueViolationField(err)
			return nil, &apperror.ConflictError{Field: field}
		}
		return nil, fmt.Errorf("failed to update user %v from database: %w", id, err)
	}

	return toGraphQLModel(user), nil
}

func (s *ServiceImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.queries.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "user"}
		}
		return fmt.Errorf("failed to delete user %v from database: %w", id, err)
	}

	return nil
}

func toGraphQLModel(u sqlc.User) *model.User {
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Cpf:       u.Cpf,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	}
}
