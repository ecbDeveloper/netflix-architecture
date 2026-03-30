package episode

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
	if strings.TrimSpace(input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title is required"}
	}
	if input.Season <= 0 {
		return nil, &apperror.ValidationError{Field: "season", Message: "season must be greater than zero"}
	}
	if input.EpisodeNumber <= 0 {
		return nil, &apperror.ValidationError{Field: "episodeNumber", Message: "episode number must be greater than zero"}
	}
	if input.DurationMinutes <= 0 {
		return nil, &apperror.ValidationError{Field: "durationMinutes", Message: "duration must be greater than zero"}
	}

	episodeID := uuid.New()

	ep, err := s.Queries.CreateEpisode(ctx, sqlc.CreateEpisodeParams{
		ID:       episodeID,
		SeriesID: input.SeriesID, Season: input.Season,
		EpisodeNumber:   input.EpisodeNumber,
		Title:           input.Title,
		DurationMinutes: input.DurationMinutes,
	})
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "episode (season + number) in this series"}
		}
		return nil, fmt.Errorf("failed to insert episode on database: %w", err)
	}

	return toGraphQLModel(ep), nil
}

func (s *Service) GetEpisode(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	ep, err := s.Queries.GetEpisode(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "episode"}
		}
		return nil, fmt.Errorf("failed to fetch episode %v from database: %w", id, err)
	}

	return toGraphQLModel(ep), nil
}

func (s *Service) ListEpisodes(ctx context.Context, seriesID int32) ([]*model.Episode, error) {
	episodes, err := s.Queries.ListEpisodesBySerie(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all episodes from series %v from database: %w", seriesID, err)
	}

	result := make([]*model.Episode, len(episodes))
	for i, ep := range episodes {
		result[i] = toGraphQLModel(ep)
	}
	return result, nil
}

func (s *Service) UpdateEpisode(ctx context.Context, id uuid.UUID, input model.UpdateEpisodeInput) (*model.Episode, error) {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title cannot be empty"}
	}
	if input.Season != nil && *input.Season <= 0 {
		return nil, &apperror.ValidationError{Field: "season", Message: "season must be greater than zero"}
	}
	if input.EpisodeNumber != nil && *input.EpisodeNumber <= 0 {
		return nil, &apperror.ValidationError{Field: "episodeNumber", Message: "episode number must be greater than zero"}
	}
	if input.DurationMinutes != nil && *input.DurationMinutes <= 0 {
		return nil, &apperror.ValidationError{Field: "durationMinutes", Message: "duration must be greater than zero"}
	}

	current, err := s.Queries.GetEpisode(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "episode"}
		}
		return nil, fmt.Errorf("failed to get episode %v to update from database: %w", id, err)
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
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "episode (season + number) in this series"}
		}
		return nil, fmt.Errorf("failed to update episode %v from database: %w", id, err)
	}

	return toGraphQLModel(ep), nil
}

func (s *Service) DeleteEpisode(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteEpisode(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "episode"}
		}
		return fmt.Errorf("failed to delete episode %v from database: %w", id, err)
	}

	return nil
}

func toGraphQLModel(e sqlc.Episode) *model.Episode {
	return &model.Episode{
		ID:              e.ID.String(),
		SeriesID:        e.SeriesID,
		Season:          e.Season,
		EpisodeNumber:   e.EpisodeNumber,
		Title:           e.Title,
		DurationMinutes: e.DurationMinutes,
		CreatedAt:       e.CreatedAt.String(),
	}
}
