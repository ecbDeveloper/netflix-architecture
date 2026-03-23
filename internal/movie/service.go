package movie

import (
	"context"
	"fmt"
	"time"

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

func (s *Service) CreateMovie(ctx context.Context, input model.CreateMovieInput) (*model.Movie, error) {
	movieID := uuid.New()

	releaseDate, err := time.Parse("2006-01-02", input.ReleaseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid release date format: %w", err)
	}

	movie, err := s.Queries.CreateMovie(ctx, sqlc.CreateMovieParams{
		ID:              movieID,
		Title:           input.Title,
		Description:     input.Description,
		DurationMinutes: input.DurationMinutes,
		ReleaseDate: pgtype.Date{
			Time:  releaseDate,
			Valid: true,
		},
		MaturityRating: input.MaturityRating,
		ContentUrl:     input.ContentURL,
	})
	if err != nil {
		return nil, err
	}

	return toGraphQLModel(movie), nil
}

func (s *Service) GetMovie(ctx context.Context, id uuid.UUID) (*model.Movie, error) {
	movie, err := s.Queries.GetMovie(ctx, id)
	if err != nil {
		return nil, err
	}

	return toGraphQLModel(movie), nil
}

func (s *Service) ListMovies(ctx context.Context) ([]*model.Movie, error) {
	movies, err := s.Queries.ListMovies(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Movie, len(movies))
	for i, m := range movies {
		result[i] = toGraphQLModel(m)
	}
	return result, nil
}

func (s *Service) UpdateMovie(ctx context.Context, id uuid.UUID, input model.UpdateMovieInput) (*model.Movie, error) {
	current, err := s.Queries.GetMovie(ctx, id)
	if err != nil {
		return nil, err
	}

	params := sqlc.UpdateMovieParams{
		ID:              id,
		Title:           current.Title,
		Description:     current.Description,
		DurationMinutes: current.DurationMinutes,
		ReleaseDate:     current.ReleaseDate,
		MaturityRating:  current.MaturityRating,
		ContentUrl:      current.ContentUrl,
	}

	if input.Title != nil {
		params.Title = *input.Title
	}
	if input.Description != nil {
		params.Description = *input.Description
	}
	if input.DurationMinutes != nil {
		params.DurationMinutes = *input.DurationMinutes
	}
	if input.ReleaseDate != nil {
		releaseDate, err := time.Parse("2006-01-02", *input.ReleaseDate)
		if err != nil {
			return nil, fmt.Errorf("invalid release date format: %w", err)
		}
		params.ReleaseDate = pgtype.Date{Time: releaseDate, Valid: true}
	}
	if input.MaturityRating != nil {
		params.MaturityRating = *input.MaturityRating
	}
	if input.ContentURL != nil {
		params.ContentUrl = *input.ContentURL
	}

	movie, err := s.Queries.UpdateMovie(ctx, params)
	if err != nil {
		return nil, err
	}

	return toGraphQLModel(movie), nil
}

func (s *Service) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	return s.Queries.DeleteMovie(ctx, id)
}

func toGraphQLModel(m sqlc.Movie) *model.Movie {
	releaseDateStr := ""
	if m.ReleaseDate.Valid {
		releaseDateStr = m.ReleaseDate.Time.Format("2006-01-02")
	}

	return &model.Movie{
		ID:              m.ID.String(),
		Title:           m.Title,
		Description:     m.Description,
		DurationMinutes: m.DurationMinutes,
		ReleaseDate:     releaseDateStr,
		MaturityRating:  m.MaturityRating,
		ContentURL:      m.ContentUrl,
	}
}
