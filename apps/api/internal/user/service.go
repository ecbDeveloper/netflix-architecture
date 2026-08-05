package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*sqlc.User, error)
	ListUsers(ctx context.Context) ([]*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type ServiceImpl struct {
	repo Repository
	pool shared.DBPool
}

func NewService(repo Repository, pool shared.DBPool) Service {
	return &ServiceImpl{
		repo: repo,
		pool: pool,
	}
}

func (s *ServiceImpl) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	if _, err := NewEmail(input.Email); err != nil {
		return nil, &apperror.ValidationError{Field: "email", Message: err.Error()}
	}
	if _, err := NewCPF(input.Cpf); err != nil {
		return nil, &apperror.ValidationError{Field: "cpf", Message: err.Error()}
	}
	if _, err := NewRawPassword(input.Password); err != nil {
		return nil, &apperror.ValidationError{Field: "password", Message: err.Error()}
	}
	if input.Name == "" {
		return nil, &apperror.ValidationError{Field: "name", Message: "name is required"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	userID := uuid.New()

	user, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
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

	normalProfileID := uuid.New()

	_, err = qtx.CreateProfile(ctx, sqlc.CreateProfileParams{
		ID:                  normalProfileID,
		UserID:              userID,
		Name:                input.Name,
		HasParentalControls: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert user normal profile on database: %w", err)
	}

	kidsProfileID := uuid.New()

	_, err = qtx.CreateProfile(ctx, sqlc.CreateProfileParams{
		ID:                  kidsProfileID,
		UserID:              userID,
		Name:                "Kids",
		HasParentalControls: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert user kids profile on database: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit user creation transaction: %w", err)
	}

	return toGraphQLModel(user), nil
}

func (s *ServiceImpl) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "user"}
		}
		return nil, fmt.Errorf("failed to fetch user %v from database: %w", id, err)
	}

	return toGraphQLModel(user), nil
}

func (s *ServiceImpl) ListUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all users from database: %w", err)
	}

	modelUsers := make([]*model.User, len(users))
	for i, u := range users {
		_ = toUserEntity(u) // maps to domain entity; methods like IsAdmin() available for domain rules
		modelUsers[i] = toGraphQLModel(u)
	}
	return modelUsers, nil
}

func (s *ServiceImpl) UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	storedUser, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "user"}
		}
		return nil, fmt.Errorf("failed to fetch user %v from database: %w", id, err)
	}

	if input.Password != nil {
		if _, err := NewRawPassword(*input.Password); err != nil {
			return nil, &apperror.ValidationError{Field: "password", Message: err.Error()}
		}
	}

	updateParams := sqlc.UpdateUserParams{
		ID:       id,
		Email:    storedUser.Email,
		Name:     storedUser.Name,
		Password: storedUser.Password,
		Cpf:      storedUser.Cpf,
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

	user, err := s.repo.UpdateUser(ctx, updateParams)
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
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "user"}
		}
		return fmt.Errorf("failed to delete user %v from database: %w", id, err)
	}

	return nil
}

func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "user"}
		}
		return nil, fmt.Errorf("failed to fetch user %v from database: %w", user.ID, err)
	}

	return &user, nil
}
