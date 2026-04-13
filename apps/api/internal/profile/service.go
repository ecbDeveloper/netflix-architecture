package profile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	CreateProfile(ctx context.Context, input model.CreateProfileInput, userID uuid.UUID) (*model.Profile, error)
	GetProfile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Profile, error)
	ListProfilesByUser(ctx context.Context, userID uuid.UUID) ([]*model.Profile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, input model.UpdateProfileInput, userID uuid.UUID) (*model.Profile, error)
	DeleteProfile(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type ServiceImpl struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) Service {
	return &ServiceImpl{
		queries: queries,
	}
}

func (s *ServiceImpl) CreateProfile(ctx context.Context, input model.CreateProfileInput, userID uuid.UUID) (*model.Profile, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, &apperror.ValidationError{Field: "name", Message: "profile name is required"}
	}

	profileID := uuid.New()

	hasParentalControls := false
	if input.HasParentalControls != nil {
		hasParentalControls = *input.HasParentalControls
	}

	p, err := s.queries.CreateProfile(ctx, sqlc.CreateProfileParams{
		ID:                  profileID,
		UserID:              userID,
		Name:                input.Name,
		HasParentalControls: hasParentalControls,
	})
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "profile name for this user"}
		}
		return nil, fmt.Errorf("failed to insert profile on database: %w", err)
	}

	return toGraphQLModel(p), nil
}

func (s *ServiceImpl) GetProfile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Profile, error) {
	p, err := s.queries.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "profile"}
		}
		return nil, fmt.Errorf("failed to fetch profile %v from database: %w", id, err)
	}

	if p.UserID != userID {
		return nil, &apperror.ForbiddenError{Message: "you can't see profiles that's not yours"}
	}

	return toGraphQLModel(p), nil
}

func (s *ServiceImpl) ListProfilesByUser(ctx context.Context, userID uuid.UUID) ([]*model.Profile, error) {
	profiles, err := s.queries.ListProfilesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all profiles from database: %w", err)
	}

	result := make([]*model.Profile, len(profiles))
	for i, p := range profiles {
		result[i] = toGraphQLModel(p)
	}
	return result, nil
}

func (s *ServiceImpl) UpdateProfile(ctx context.Context, id uuid.UUID, input model.UpdateProfileInput, userID uuid.UUID) (*model.Profile, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return nil, &apperror.ValidationError{Field: "name", Message: "profile name cannot be empty"}
	}

	current, err := s.queries.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "profile"}
		}
		return nil, fmt.Errorf("failed to get profile %v to update from database: %w", id, err)
	}

	if current.UserID != userID {
		return nil, &apperror.ForbiddenError{Message: "you can't update profiles that's not yours"}
	}

	params := sqlc.UpdateProfileParams{
		ID:                  id,
		Name:                current.Name,
		HasParentalControls: current.HasParentalControls,
	}

	if input.Name != nil {
		params.Name = *input.Name
	}
	if input.HasParentalControls != nil {
		params.HasParentalControls = *input.HasParentalControls
	}

	p, err := s.queries.UpdateProfile(ctx, params)
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "profile name for this user"}
		}
		return nil, fmt.Errorf("failed to update profile %v from database: %w", id, err)
	}

	return toGraphQLModel(p), nil
}

func (s *ServiceImpl) DeleteProfile(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	current, err := s.queries.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "profile"}
		}
		return fmt.Errorf("failed to get profile %v to update from database: %w", id, err)
	}

	if current.UserID != userID {
		return &apperror.ForbiddenError{Message: "you can't delete profiles that's not yours"}
	}

	if err := s.queries.DeleteProfile(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "profile"}
		}
		return fmt.Errorf("failed to delete profile %v from database: %w", id, err)
	}
	return nil
}

func toGraphQLModel(p sqlc.Profile) *model.Profile {
	return &model.Profile{
		ID:                  p.ID,
		UserID:              p.UserID,
		Name:                p.Name,
		HasParentalControls: p.HasParentalControls,
		CreatedAt:           p.CreatedAt.String(),
		UpdatedAt:           p.UpdatedAt.String(),
	}
}
