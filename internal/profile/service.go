package profile

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	Queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{
		Queries: queries,
	}
}

func (s *Service) CreateProfile(ctx context.Context, input model.CreateProfileInput) (*model.Profile, error) {
	profileID := uuid.New()

	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}

	hasParentalControls := false
	if input.HasParentalControls != nil {
		hasParentalControls = *input.HasParentalControls
	}

	p, err := s.Queries.CreateProfile(ctx, sqlc.CreateProfileParams{
		ID: profileID,
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		Name:                input.Name,
		HasParentalControls: hasParentalControls,
	})
	if err != nil {
		return nil, err
	}

	return toGraphQLModel(p), nil
}

func (s *Service) GetProfile(ctx context.Context, id uuid.UUID) (*model.Profile, error) {
	p, err := s.Queries.GetProfile(ctx, id)
	if err != nil {
		return nil, err
	}

	return toGraphQLModel(p), nil
}

func (s *Service) ListProfiles(ctx context.Context, userID uuid.UUID) ([]*model.Profile, error) {
	profiles, err := s.Queries.ListProfilesByUser(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*model.Profile, len(profiles))
	for i, p := range profiles {
		result[i] = toGraphQLModel(p)
	}
	return result, nil
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, input model.UpdateProfileInput) (*model.Profile, error) {
	current, err := s.Queries.GetProfile(ctx, id)
	if err != nil {
		return nil, err
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

	p, err := s.Queries.UpdateProfile(ctx, params)
	if err != nil {
		return nil, err
	}

	return toGraphQLModel(p), nil
}

func (s *Service) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	return s.Queries.DeleteProfile(ctx, id)
}

func toGraphQLModel(p sqlc.Profile) *model.Profile {
	userIDStr := ""
	if p.UserID.Valid {
		userIDStr = uuid.UUID(p.UserID.Bytes).String()
	}

	return &model.Profile{
		ID:                  p.ID.String(),
		UserID:              userIDStr,
		Name:                p.Name,
		HasParentalControls: p.HasParentalControls,
		CreatedAt:           p.CreatedAt.String(),
		UpdatedAt:           p.UpdatedAt.String(),
	}
}
