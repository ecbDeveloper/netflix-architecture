package review

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

type Service interface {
	CreateReview(ctx context.Context, input model.CreateReviewInput, profileID uuid.UUID) (*model.Review, error)
	GetReview(ctx context.Context, id uuid.UUID) (*model.Review, error)
	ListReviews(ctx context.Context, profileID uuid.UUID) ([]*model.Review, error)
	UpdateReview(ctx context.Context, id uuid.UUID, input model.UpdateReviewInput, profileID uuid.UUID) (*model.Review, error)
	DeleteReview(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error
}

type ServiceImpl struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) Service {
	return &ServiceImpl{
		queries: queries,
	}
}

func (s *ServiceImpl) CreateReview(ctx context.Context, input model.CreateReviewInput, profileID uuid.UUID) (*model.Review, error) {
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

	params.ProfileID = profileID

	if input.MovieID != nil {
		params.MovieID = pgtype.UUID{Bytes: *input.MovieID, Valid: true}
	}

	if input.EpisodeID != nil {
		params.EpisodeID = pgtype.UUID{Bytes: *input.EpisodeID, Valid: true}
	}

	if input.Comment != nil {
		params.Comment = pgtype.Text{String: *input.Comment, Valid: true}
	}

	r, err := s.queries.CreateReview(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to insert review on database: %w", err)
	}

	return toGraphQLModel(r), nil
}

func (s *ServiceImpl) GetReview(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	r, err := s.queries.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "review"}
		}
		return nil, fmt.Errorf("failed to fetch review %v from database: %w", id, err)
	}

	return toGraphQLModel(r), nil
}

func (s *ServiceImpl) ListReviewsByProfile(ctx context.Context, profileID uuid.UUID) ([]*model.Review, error) {
	reviews, err := s.queries.ListReviewsByProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all reviews from database: %w", err)
	}

	result := make([]*model.Review, len(reviews))
	for i, r := range reviews {
		result[i] = toGraphQLModel(r)
	}
	return result, nil
}

func (s *ServiceImpl) UpdateReview(ctx context.Context, id uuid.UUID, input model.UpdateReviewInput, profileID uuid.UUID) (*model.Review, error) {
	if input.Rating != nil && (*input.Rating < 1 || *input.Rating > 5) {
		return nil, &apperror.ValidationError{Field: "rating", Message: "rating must be between 1 and 5"}
	}

	current, err := s.queries.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "review"}
		}
		return nil, fmt.Errorf("failed to get review %v to update from database: %w", id, err)
	}

	if current.ProfileID != profileID {
		return nil, &apperror.ForbiddenError{Message: "you can't update reviews that's not yours"}
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

	r, err := s.queries.UpdateReview(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update review %v from database: %w", id, err)
	}

	return toGraphQLModel(r), nil
}

func (s *ServiceImpl) DeleteReview(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error {
	current, err := s.queries.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "review"}
		}
		return fmt.Errorf("failed to get review %v to update from database: %w", id, err)
	}

	if current.ProfileID != profileID {
		return &apperror.ForbiddenError{Message: "you can't delete reviews that's not yours"}
	}

	if err := s.queries.DeleteReview(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "review"}
		}
		return fmt.Errorf("failed to delete review %v from database: %w", id, err)
	}
	return nil
}

func toGraphQLModel(r sqlc.Review) *model.Review {
	m := &model.Review{
		ID:     r.ID,
		Rating: r.Rating,
	}

	if r.ProfileID != uuid.Nil {
		m.ProfileID = r.ProfileID
	}
	if r.MovieID.Valid {
		movieID, _ := uuid.Parse(r.MovieID.String())
		m.MovieID = &movieID
	}
	if r.EpisodeID.Valid {
		episodeID, _ := uuid.Parse(r.EpisodeID.String())
		m.EpisodeID = &episodeID
	}
	if r.Comment.Valid {
		m.Comment = &r.Comment.String
	}
	if r.CreatedAt.Valid {
		m.CreatedAt = r.CreatedAt.Time.String()
	}

	return m
}
