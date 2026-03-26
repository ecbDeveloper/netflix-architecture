package series

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model"
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

func (s *Service) CreateSeries(ctx context.Context, input model.CreateSeriesInput) (*model.Series, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title is required"}
	}
	if input.Description == nil || strings.TrimSpace(*input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description is required"}
	}
	if input.ReleaseDate == nil || strings.TrimSpace(*input.ReleaseDate) == "" {
		return nil, &apperror.ValidationError{Field: "releaseDate", Message: "release date is required"}
	}
	if input.MaturityRating == nil || strings.TrimSpace(*input.MaturityRating) == "" {
		return nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturity rating is required"}
	}

	params := sqlc.CreateSerieParams{
		Title:       input.Title,
		Description: pgtype.Text{String: *input.Description, Valid: true},
		MaturityRating: pgtype.Text{String: *input.MaturityRating, Valid: true},
	}

	releaseDate, err := time.Parse("2006-01-02", *input.ReleaseDate)
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

func (s *Service) GetSeries(ctx context.Context, id int32) (*model.Series, error) {
	serie, err := s.Queries.GetSerie(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "series"}
		}
		return nil, fmt.Errorf("failed to fetch series %v from database: %w", id, err)
	}

	return toGraphQLModel(serie), nil
}

func (s *Service) ListSeries(ctx context.Context) ([]*model.Series, error) {
	seriesList, err := s.Queries.ListSeries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all series from database: %w", err)
	}

	result := make([]*model.Series, len(seriesList))
	for i, serie := range seriesList {
		result[i] = toGraphQLModel(serie)
	}
	return result, nil
}

func (s *Service) UpdateSeries(ctx context.Context, id int32, input model.UpdateSeriesInput) (*model.Series, error) {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title cannot be empty"}
	}
	if input.Description != nil && strings.TrimSpace(*input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description cannot be empty"}
	}
	if input.MaturityRating != nil && strings.TrimSpace(*input.MaturityRating) == "" {
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
	}

	if input.Title != nil {
		params.Title = *input.Title
	}
	if input.Description != nil {
		params.Description = pgtype.Text{String: *input.Description, Valid: true}
	}
	if input.ReleaseDate != nil {
		releaseDate, err := time.Parse("2006-01-02", *input.ReleaseDate)
		if err != nil {
			return nil, &apperror.ValidationError{Field: "releaseDate", Message: "invalid date format, use YYYY-MM-DD"}
		}
		params.ReleaseDate = pgtype.Date{Time: releaseDate, Valid: true}
	}
	if input.MaturityRating != nil {
		params.MaturityRating = pgtype.Text{String: *input.MaturityRating, Valid: true}
	}

	serie, err := s.Queries.UpdateSerie(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update series %v from database: %w", id, err)
	}

	return toGraphQLModel(serie), nil
}

func (s *Service) DeleteSeries(ctx context.Context, id int32) error {
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
		ID:    strconv.Itoa(int(s.ID)),
		Title: s.Title,
	}

	if s.Description.Valid {
		m.Description = &s.Description.String
	}
	if s.ReleaseDate.Valid {
		rd := s.ReleaseDate.Time.Format("2006-01-02")
		m.ReleaseDate = &rd
	}
	if s.MaturityRating.Valid {
		m.MaturityRating = &s.MaturityRating.String
	}

	return m
}
