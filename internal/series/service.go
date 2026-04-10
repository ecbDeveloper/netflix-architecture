package series

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

type Service interface {
	CreateSeries(ctx context.Context, input model.CreateSeriesInput) (*model.Series, error)
	GetSeries(ctx context.Context, id uuid.UUID, profileID uuid.UUID) (*model.Series, error)
	ListSeries(ctx context.Context, profileID uuid.UUID) ([]*model.Series, error)
	UpdateSeries(ctx context.Context, id uuid.UUID, input model.UpdateSeriesInput) (*model.Series, error)
	DeleteSeries(ctx context.Context, id uuid.UUID) error
}

type ServiceImpl struct {
	Queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) Service {
	return &ServiceImpl{
		Queries: queries,
	}
}

func (s *ServiceImpl) CreateSeries(ctx context.Context, input model.CreateSeriesInput) (*model.Series, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title is required"}
	}
	if strings.TrimSpace(input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description is required"}
	}
	if strings.TrimSpace(input.ReleaseDate) == "" {
		return nil, &apperror.ValidationError{Field: "releaseDate", Message: "release date is required"}
	}
	if strings.TrimSpace(string(input.MaturityRating)) == "" {
		return nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturity rating is required"}
	}

	params := sqlc.CreateSerieParams{
		Title:          input.Title,
		Description:    input.Description,
		MaturityRating: sqlc.MaturityRating(input.MaturityRating),
		GenreID:        input.GenreID,
	}

	releaseDate, err := time.Parse("2006-01-02", input.ReleaseDate)
	if err != nil {
		return nil, &apperror.ValidationError{Field: "releaseDate", Message: "invalid date format, use YYYY-MM-DD"}
	}
	params.ReleaseDate = pgtype.Date{Time: releaseDate, Valid: true}

	serie, err := s.Queries.CreateSerie(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to insert series on database: %w", err)
	}

	return toGraphQLModel(serie), nil
}

func (s *ServiceImpl) GetSeries(ctx context.Context, id uuid.UUID, profileID uuid.UUID) (*model.Series, error) {
	profile, err := s.Queries.GetProfile(ctx, profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "profile"}
		}
		return nil, fmt.Errorf("failed to get profile %v from database: %w", profileID, err)
	}

	series, err := s.Queries.GetSerie(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "series"}
		}
		return nil, fmt.Errorf("failed to fetch series %v from database: %w", id, err)
	}

	if series.MaturityRating != sqlc.MaturityRatingL && profile.HasParentalControls {
		return nil, &apperror.ForbiddenError{Message: "this profile cannot access this content due to parental controls"}
	}

	return toGraphQLModel(series), nil
}

func (s *ServiceImpl) ListSeries(ctx context.Context, profileID uuid.UUID) ([]*model.Series, error) {
	profile, err := s.Queries.GetProfile(ctx, profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "profile"}
		}
		return nil, fmt.Errorf("failed to get profile %v from database: %w", profileID, err)
	}

	var seriesList []sqlc.Series
	if profile.HasParentalControls {
		seriesList, err = s.Queries.ListKidsSeries(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all series from database: %w", err)
		}
	} else {

		seriesList, err = s.Queries.ListSeries(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all series from database: %w", err)
		}
	}

	result := make([]*model.Series, len(seriesList))
	for i, serie := range seriesList {
		result[i] = toGraphQLModel(serie)
	}
	return result, nil
}

func (s *ServiceImpl) UpdateSeries(ctx context.Context, id uuid.UUID, input model.UpdateSeriesInput) (*model.Series, error) {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title cannot be empty"}
	}
	if input.Description != nil && strings.TrimSpace(*input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description cannot be empty"}
	}
	if input.MaturityRating != nil && strings.TrimSpace(string(*input.MaturityRating)) == "" {
		return nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturity rating cannot be empty"}
	}

	current, err := s.Queries.GetSerie(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "series"}
		}
		return nil, fmt.Errorf("failed to get series %v to update from database: %w", id, err)
	}

	params := sqlc.UpdateSerieParams{
		ID:             id,
		Title:          current.Title,
		Description:    current.Description,
		ReleaseDate:    current.ReleaseDate,
		MaturityRating: current.MaturityRating,
		GenreID:        current.GenreID,
	}

	if input.Title != nil {
		params.Title = *input.Title
	}
	if input.Description != nil {
		params.Description = *input.Description
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
	if input.GenreID != nil {
		params.GenreID = *input.GenreID
	}

	serie, err := s.Queries.UpdateSerie(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update series %v from database: %w", id, err)
	}

	return toGraphQLModel(serie), nil
}

func (s *ServiceImpl) DeleteSeries(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteSerie(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "series"}
		}
		return fmt.Errorf("failed to delete series %v from database: %w", id, err)
	}
	return nil
}

func toGraphQLModel(s sqlc.Series) *model.Series {
	m := &model.Series{
		ID:             s.ID.String(),
		Title:          s.Title,
		Description:    s.Description,
		MaturityRating: model.MaturityRating(s.MaturityRating),
		GenreID:        s.GenreID,
	}

	if s.ReleaseDate.Valid {
		m.ReleaseDate = s.ReleaseDate.Time.Format("2006-01-02")
	}

	return m
}
