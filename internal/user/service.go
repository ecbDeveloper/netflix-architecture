package user

import (
	"context"

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
	salt := uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       userID,
		Email:    input.Email,
		Name:     input.Name,
		Cpf:      input.Cpf,
		Password: string(hashedPassword),
		Salt:     salt,
	})
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Cpf:   user.Cpf,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.Queries.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Cpf:       user.Cpf,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.Queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	modelUsers := make([]*model.User, len(users))
	for i, user := range users {
		modelUsers[i] = &model.User{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			Cpf:       user.Cpf,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		}
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
		salt := uuid.New().String() // Generate new salt for new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updateParams.Password = string(hashedPassword)
		updateParams.Salt = salt
	}

	user, err := s.Queries.UpdateUser(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
	}, nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := s.Queries.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
