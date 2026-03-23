package review

import (
	"context"
	"fmt"
	"strconv"

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

func (s *Service) CreateReview(ctx context.Context, input model.CreateReviewInput) (*model.Review, error) {
	params := sqlc.CreateReviewParams{
		Rating: input.Rating,
	}

	profileUUID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}
	params.ProfileID = pgtype.UUID{Bytes: profileUUID, Valid: true}

	if input.MovieID != nil {
		movieUUID, err := uuid.Parse(*input.MovieID)
		if err != nil {
			return nil, fmt.Errorf("invalid movie ID: %w", err)
		}
		params.MovieID = pgtype.UUID{Bytes: movieUUID, Valid: true}
	}

	if input.EpisodeID != nil {
		episodeUUID, err := uuid.Parse(*input.EpisodeID)
		if err != nil {
			return nil, fmt.Errorf("invalid episode ID: %w", err)
		}
		params.EpisodeID = pgtype.UUID{Bytes: episodeUUID, Valid: true}
	}

	if input.Comment != nil {
		params.Comment = pgtype.Text{String: *input.Comment, Valid: true}
	}

	r, err := s.Queries.CreateReview(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return toGraphQLModel(r), nil
}

func (s *Service) GetReview(ctx context.Context, id int32) (*model.Review, error) {
	r, err := s.Queries.GetReview(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return toGraphQLModel(r), nil
}

func (s *Service) ListReviews(ctx context.Context, profileID uuid.UUID) ([]*model.Review, error) {
	reviews, err := s.Queries.ListReviewsByProfile(ctx, pgtype.UUID{
		Bytes: profileID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}

	result := make([]*model.Review, len(reviews))
	for i, r := range reviews {
		result[i] = toGraphQLModel(r)
	}
	return result, nil
}

func (s *Service) UpdateReview(ctx context.Context, id int32, input model.UpdateReviewInput) (*model.Review, error) {
	current, err := s.Queries.GetReview(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get review to update: %w", err)
	}

	params := sqlc.UpdateReviewParams{
		ID:      id,
		Rating:  current.Rating,
		Comment: current.Comment,
	}

	if input.Rating != nil {
		params.Rating = *input.Rating
	}
	if input.Comment != nil {
		params.Comment = pgtype.Text{String: *input.Comment, Valid: true}
	}

	r, err := s.Queries.UpdateReview(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	return toGraphQLModel(r), nil
}

func (s *Service) DeleteReview(ctx context.Context, id int32) error {
	if err := s.Queries.DeleteReview(ctx, id); err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}

func toGraphQLModel(r sqlc.Review) *model.Review {
	m := &model.Review{
		ID:     strconv.Itoa(int(r.ID)),
		Rating: r.Rating,
	}

	if r.ProfileID.Valid {
		m.ProfileID = uuid.UUID(r.ProfileID.Bytes).String()
	}
	if r.MovieID.Valid {
		mid := uuid.UUID(r.MovieID.Bytes).String()
		m.MovieID = &mid
	}
	if r.EpisodeID.Valid {
		eid := uuid.UUID(r.EpisodeID.Bytes).String()
		m.EpisodeID = &eid
	}
	if r.Comment.Valid {
		m.Comment = &r.Comment.String
	}
	if r.CreatedAt.Valid {
		m.CreatedAt = r.CreatedAt.Time.String()
	}

	return m
}
