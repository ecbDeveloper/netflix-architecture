package series

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/model"
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
	params := sqlc.CreateSerieParams{
		Title: input.Title,
	}

	if input.Description != nil {
		params.Description = pgtype.Text{String: *input.Description, Valid: true}
	}
	if input.ReleaseDate != nil {
		releaseDate, err := time.Parse("2006-01-02", *input.ReleaseDate)
		if err != nil {
			return nil, fmt.Errorf("invalid release date format: %w", err)
		}
		params.ReleaseDate = pgtype.Date{Time: releaseDate, Valid: true}
	}
	if input.MaturityRating != nil {
		params.MaturityRating = pgtype.Text{String: *input.MaturityRating, Valid: true}
	}

	serie, err := s.Queries.CreateSerie(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to insert series on database: %w", err)
	}

	return toGraphQLModel(serie), nil
}

func (s *Service) GetSeries(ctx context.Context, id int32) (*model.Series, error) {
	serie, err := s.Queries.GetSerie(ctx, id)
	if err != nil {
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
	current, err := s.Queries.GetSerie(ctx, id)
	if err != nil {
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
			return nil, fmt.Errorf("invalid release date format: %w", err)
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
