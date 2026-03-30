package movie

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (s *Service) CreateMovie(ctx context.Context, input model.CreateMovieInput) (*model.Movie, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title is required"}
	}
	if strings.TrimSpace(input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description is required"}
	}
	if input.DurationMinutes <= 0 {
		return nil, &apperror.ValidationError{Field: "durationMinutes", Message: "duration must be greater than zero"}
	}
	if strings.TrimSpace(string(input.MaturityRating)) == "" {
		return nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturity rating is required"}
	}
	if strings.TrimSpace(input.ContentURL) == "" {
		return nil, &apperror.ValidationError{Field: "contentUrl", Message: "content URL is required"}
	}

	movieID := uuid.New()

	releaseDate, err := time.Parse("2006-01-02", input.ReleaseDate)
	if err != nil {
		return nil, &apperror.ValidationError{Field: "releaseDate", Message: "invalid date format, use YYYY-MM-DD"}
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
		MaturityRating: sqlc.MaturityRating(input.MaturityRating),
		ContentUrl:     input.ContentURL,
		GenreID:        input.GenreID,
	})
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "title"}
		}
		return nil, fmt.Errorf("failed to insert movie on database: %w", err)
	}

	return toGraphQLModel(movie), nil
}

func (s *Service) GetMovie(ctx context.Context, id uuid.UUID) (*model.Movie, error) {
	movie, err := s.Queries.GetMovie(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "movie"}
		}
		return nil, fmt.Errorf("failed to fetch movie %v from database: %w", id, err)
	}

	return toGraphQLModel(movie), nil
}

func (s *Service) ListMovies(ctx context.Context) ([]*model.Movie, error) {
	movies, err := s.Queries.ListMovies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all movies from database: %w", err)
	}

	result := make([]*model.Movie, len(movies))
	for i, m := range movies {
		result[i] = toGraphQLModel(m)
	}
	return result, nil
}

func (s *Service) UpdateMovie(ctx context.Context, id uuid.UUID, input model.UpdateMovieInput) (*model.Movie, error) {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title cannot be empty"}
	}
	if input.Description != nil && strings.TrimSpace(*input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description cannot be empty"}
	}
	if input.DurationMinutes != nil && *input.DurationMinutes <= 0 {
		return nil, &apperror.ValidationError{Field: "durationMinutes", Message: "duration must be greater than zero"}
	}
	if input.MaturityRating != nil && strings.TrimSpace(string(*input.MaturityRating)) == "" {
		return nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturity rating cannot be empty"}
	}
	if input.ContentURL != nil && strings.TrimSpace(*input.ContentURL) == "" {
		return nil, &apperror.ValidationError{Field: "contentUrl", Message: "content URL cannot be empty"}
	}

	current, err := s.Queries.GetMovie(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "movie"}
		}
		return nil, fmt.Errorf("failed to get movie %v to update from database: %w", id, err)
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
			return nil, &apperror.ValidationError{Field: "releaseDate", Message: "invalid date format, use YYYY-MM-DD"}
		}
		params.ReleaseDate = pgtype.Date{Time: releaseDate, Valid: true}
	}
	if input.MaturityRating != nil {
		params.MaturityRating = sqlc.MaturityRating(*input.MaturityRating)
	}
	if input.ContentURL != nil {
		params.ContentUrl = *input.ContentURL
	}

	movie, err := s.Queries.UpdateMovie(ctx, params)
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "title"}
		}
		return nil, fmt.Errorf("failed to update movie %v from database: %w", id, err)
	}

	return toGraphQLModel(movie), nil
}

func (s *Service) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteMovie(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "movie"}
		}
		return fmt.Errorf("failed to delete movie %v from database: %w", id, err)
	}
	return nil
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
		MaturityRating:  model.MaturityRating(m.MaturityRating),
		ContentURL:      m.ContentUrl,
		GenreID:         m.GenreID,
	}
}
