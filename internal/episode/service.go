package episode

import (
	"context"
	"fmt"

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

func (s *Service) CreateEpisode(ctx context.Context, input model.CreateEpisodeInput) (*model.Episode, error) {
	episodeID := uuid.New()

	ep, err := s.Queries.CreateEpisode(ctx, sqlc.CreateEpisodeParams{
		ID: episodeID,
		SerieID: pgtype.Int4{
			Int32: input.SerieID,
			Valid: true,
		},
		Season:          input.Season,
		EpisodeNumber:   input.EpisodeNumber,
		Title:           input.Title,
		DurationMinutes: input.DurationMinutes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create episode: %w", err)
	}

	return toGraphQLModel(ep), nil
}

func (s *Service) GetEpisode(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	ep, err := s.Queries.GetEpisode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode by id: %w", err)
	}

	return toGraphQLModel(ep), nil
}

func (s *Service) ListEpisodes(ctx context.Context, serieID int32) ([]*model.Episode, error) {
	episodes, err := s.Queries.ListEpisodesBySerie(ctx, pgtype.Int4{
		Int32: serieID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list episodes: %w", err)
	}

	result := make([]*model.Episode, len(episodes))
	for i, ep := range episodes {
		result[i] = toGraphQLModel(ep)
	}
	return result, nil
}

func (s *Service) UpdateEpisode(ctx context.Context, id uuid.UUID, input model.UpdateEpisodeInput) (*model.Episode, error) {
	current, err := s.Queries.GetEpisode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode to update: %w", err)
	}

	params := sqlc.UpdateEpisodeParams{
		ID:              id,
		Season:          current.Season,
		EpisodeNumber:   current.EpisodeNumber,
		Title:           current.Title,
		DurationMinutes: current.DurationMinutes,
	}

	if input.Season != nil {
		params.Season = *input.Season
	}
	if input.EpisodeNumber != nil {
		params.EpisodeNumber = *input.EpisodeNumber
	}
	if input.Title != nil {
		params.Title = *input.Title
	}
	if input.DurationMinutes != nil {
		params.DurationMinutes = *input.DurationMinutes
	}

	ep, err := s.Queries.UpdateEpisode(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update episode: %w", err)
	}

	return toGraphQLModel(ep), nil
}

func (s *Service) DeleteEpisode(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteEpisode(ctx, id); err != nil {
		return fmt.Errorf("failed to delete episode: %w", err)
	}

	return nil
}

func toGraphQLModel(e sqlc.Episode) *model.Episode {
	return &model.Episode{
		ID:              e.ID.String(),
		SerieID:         e.SerieID.Int32,
		Season:          e.Season,
		EpisodeNumber:   e.EpisodeNumber,
		Title:           e.Title,
		DurationMinutes: e.DurationMinutes,
		CreatedAt:       e.CreatedAt.String(),
	}
}
