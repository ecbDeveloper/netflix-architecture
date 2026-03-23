package user

import (
	"context"
	"fmt"

	// For context
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model" // Assuming this is the correct import path for model_gen.go
	"github.com/google/uuid"
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

func (s *Service) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	userID := uuid.New()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       userID,
		Email:    input.Email,
		Name:     input.Name,
		Cpf:      input.Cpf,
		Password: string(hashedPassword),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return toGraphQLModel(user), nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.Queries.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return toGraphQLModel(user), nil
}

func (s *Service) ListUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.Queries.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	modelUsers := make([]*model.User, len(users))
	for i, user := range users {
		modelUsers[i] = toGraphQLModel(user)
	}
	return modelUsers, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
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

	user, err := s.Queries.UpdateUser(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return toGraphQLModel(user), nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func toGraphQLModel(u sqlc.User) *model.User {
	return &model.User{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		Cpf:       u.Cpf,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	}
}
