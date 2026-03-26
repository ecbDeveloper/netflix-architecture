package watchhistory

import (
	"context"
	"errors"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (s *Service) CreateWatchHistory(ctx context.Context, input model.CreateWatchHistoryInput) (*model.WatchHistory, error) {
	if input.MovieID == nil && input.EpisodeID == nil {
		return nil, &apperror.ValidationError{Field: "movieId/episodeId", Message: "movie or episode is required"}
	}
	if input.MovieID != nil && input.EpisodeID != nil {
		return nil, &apperror.ValidationError{Field: "movieId/episodeId", Message: "provide only movie or episode, not both"}
	}

	whID := uuid.New()

	params := sqlc.CreateWatchHistoryParams{
		ID: whID,
	}

	profileUUID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, &apperror.ValidationError{Field: "profileId", Message: "invalid profile ID"}
	}
	params.ProfileID = pgtype.UUID{Bytes: profileUUID, Valid: true}

	if input.MovieID != nil {
		movieUUID, err := uuid.Parse(*input.MovieID)
		if err != nil {
			return nil, &apperror.ValidationError{Field: "movieId", Message: "invalid movie ID"}
		}
		params.MovieID = pgtype.UUID{Bytes: movieUUID, Valid: true}
	}

	if input.EpisodeID != nil {
		episodeUUID, err := uuid.Parse(*input.EpisodeID)
		if err != nil {
			return nil, &apperror.ValidationError{Field: "episodeId", Message: "invalid episode ID"}
		}
		params.EpisodeID = pgtype.UUID{Bytes: episodeUUID, Valid: true}
	}

	if input.LastPositionSeconds != nil {
		params.LastPositionSeconds = pgtype.Int4{Int32: *input.LastPositionSeconds, Valid: true}
	}

	if input.IsCompleted != nil {
		params.IsCompleted = pgtype.Bool{Bool: *input.IsCompleted, Valid: true}
	}

	wh, err := s.Queries.CreateWatchHistory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to insert watch history on database: %w", err)
	}

	return toGraphQLModel(wh), nil
}

func (s *Service) GetWatchHistory(ctx context.Context, id uuid.UUID) (*model.WatchHistory, error) {
	wh, err := s.Queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "watch history"}
		}
		return nil, fmt.Errorf("failed to fetch watch history %v from database: %w", id, err)
	}

	return toGraphQLModel(wh), nil
}

func (s *Service) ListWatchHistories(ctx context.Context, profileID uuid.UUID) ([]*model.WatchHistory, error) {
	histories, err := s.Queries.ListWatchHistoryByProfile(ctx, pgtype.UUID{
		Bytes: profileID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all watch histories from database: %w", err)
	}

	result := make([]*model.WatchHistory, len(histories))
	for i, wh := range histories {
		result[i] = toGraphQLModel(wh)
	}
	return result, nil
}

func (s *Service) UpdateWatchHistory(ctx context.Context, id uuid.UUID, input model.UpdateWatchHistoryInput) (*model.WatchHistory, error) {
	current, err := s.Queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "watch history"}
		}
		return nil, fmt.Errorf("failed to get watch history %v to update from database: %w", id, err)
	}

	params := sqlc.UpdateWatchProgressParams{
		ID:                  id,
		LastPositionSeconds: current.LastPositionSeconds,
		IsCompleted:         current.IsCompleted,
	}

	if input.LastPositionSeconds != nil {
		params.LastPositionSeconds = pgtype.Int4{Int32: *input.LastPositionSeconds, Valid: true}
	}
	if input.IsCompleted != nil {
		params.IsCompleted = pgtype.Bool{Bool: *input.IsCompleted, Valid: true}
	}

	wh, err := s.Queries.UpdateWatchProgress(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update watch history %v from database: %w", id, err)
	}

	return toGraphQLModel(wh), nil
}

func (s *Service) DeleteWatchHistory(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteWatchHistory(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "watch history"}
		}
		return fmt.Errorf("failed to delete watch history %v from database: %w", id, err)
	}
	return nil
}

func toGraphQLModel(wh sqlc.WatchHistory) *model.WatchHistory {
	m := &model.WatchHistory{
		ID: wh.ID.String(),
	}

	if wh.ProfileID.Valid {
		m.ProfileID = uuid.UUID(wh.ProfileID.Bytes).String()
	}
	if wh.MovieID.Valid {
		mid := uuid.UUID(wh.MovieID.Bytes).String()
		m.MovieID = &mid
	}
	if wh.EpisodeID.Valid {
		eid := uuid.UUID(wh.EpisodeID.Bytes).String()
		m.EpisodeID = &eid
	}
	if wh.WatchedAt.Valid {
		m.WatchedAt = wh.WatchedAt.Time.String()
	}
	if wh.LastPositionSeconds.Valid {
		lps := wh.LastPositionSeconds.Int32
		m.LastPositionSeconds = &lps
	}
	if wh.IsCompleted.Valid {
		ic := wh.IsCompleted.Bool
		m.IsCompleted = &ic
	}

	return m
}
