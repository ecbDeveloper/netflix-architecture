package review

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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

func (s *Service) CreateReview(ctx context.Context, input model.CreateReviewInput) (*model.Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, &apperror.ValidationError{Field: "rating", Message: "rating must be between 1 and 5"}
	}
	if input.MovieID == nil && input.EpisodeID == nil {
		return nil, &apperror.ValidationError{Field: "movieId/episodeId", Message: "movie or episode is required for the review"}
	}
	if input.MovieID != nil && input.EpisodeID != nil {
		return nil, &apperror.ValidationError{Field: "movieId/episodeId", Message: "provide only movie or episode, not both"}
	}

	params := sqlc.CreateReviewParams{
		Rating: input.Rating,
	}

	profileUUID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, &apperror.ValidationError{Field: "profileId", Message: "invalid profile ID"}
	}
	params.ProfileID = profileUUID

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

	if input.Comment != nil {
		params.Comment = pgtype.Text{String: *input.Comment, Valid: true}
	}

	r, err := s.Queries.CreateReview(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to insert review on database: %w", err)
	}

	return toGraphQLModel(r), nil
}

func (s *Service) GetReview(ctx context.Context, id int32) (*model.Review, error) {
	r, err := s.Queries.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "review"}
		}
		return nil, fmt.Errorf("failed to fetch review %v from database: %w", id, err)
	}

	return toGraphQLModel(r), nil
}

func (s *Service) ListReviews(ctx context.Context, profileID uuid.UUID) ([]*model.Review, error) {
	reviews, err := s.Queries.ListReviewsByProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all reviews from database: %w", err)
	}

	result := make([]*model.Review, len(reviews))
	for i, r := range reviews {
		result[i] = toGraphQLModel(r)
	}
	return result, nil
}

func (s *Service) UpdateReview(ctx context.Context, id int32, input model.UpdateReviewInput) (*model.Review, error) {
	if input.Rating != nil && (*input.Rating < 1 || *input.Rating > 5) {
		return nil, &apperror.ValidationError{Field: "rating", Message: "rating must be between 1 and 5"}
	}

	current, err := s.Queries.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "review"}
		}
		return nil, fmt.Errorf("failed to get review %v to update from database: %w", id, err)
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
		return nil, fmt.Errorf("failed to update review %v from database: %w", id, err)
	}

	return toGraphQLModel(r), nil
}

func (s *Service) DeleteReview(ctx context.Context, id int32) error {
	if err := s.Queries.DeleteReview(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "review"}
		}
		return fmt.Errorf("failed to delete review %v from database: %w", id, err)
	}
	return nil
}

func toGraphQLModel(r sqlc.Review) *model.Review {
	m := &model.Review{
		ID:     strconv.Itoa(int(r.ID)),
		Rating: r.Rating,
	}

	if r.ProfileID != uuid.Nil {
		m.ProfileID = r.ProfileID.String()
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
