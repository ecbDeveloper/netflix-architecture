package review

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateReview(ctx context.Context, input model.CreateReviewInput, profileID uuid.UUID) (*model.Review, error)
	GetReview(ctx context.Context, id uuid.UUID) (*model.Review, error)
	ListReviewsByProfile(ctx context.Context, profileID uuid.UUID) ([]*model.Review, error)
	ListReviewsByEpisode(ctx context.Context, episodeID uuid.UUID, profileID uuid.UUID, userID uuid.UUID) ([]*model.Review, error)
	ListReviewsByMovie(ctx context.Context, movieID uuid.UUID) ([]*model.Review, error)
	UpdateReview(ctx context.Context, id uuid.UUID, input model.UpdateReviewInput, profileID uuid.UUID) (*model.Review, error)
	DeleteReview(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error
}

type ServiceImpl struct {
	repo           Repository
	episodeService episode.Service
}

func NewService(repo Repository, es episode.Service) Service {
	return &ServiceImpl{
		repo:           repo,
		episodeService: es,
	}
}

func (s *ServiceImpl) CreateReview(ctx context.Context, input model.CreateReviewInput, profileID uuid.UUID) (*model.Review, error) {
	if _, err := NewRating(input.Rating); err != nil {
		return nil, &apperror.ValidationError{Field: "rating", Message: err.Error()}
	}
	if input.MovieID == nil && input.EpisodeID == nil {
		return nil, &apperror.ValidationError{Field: "movieId/episodeId", Message: "movie or episode is required for the review"}
	}
	if input.MovieID != nil && input.EpisodeID != nil {
		return nil, &apperror.ValidationError{Field: "movieId/episodeId", Message: "provide only movie or episode, not both"}
	}

	profileReviews, err := s.ListReviewsByProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews from profile %v to create a review: %w", profileID, err)
	}

	for _, review := range profileReviews {
		if input.MovieID != nil && review.MovieID == *input.MovieID {
			return nil, &apperror.UnprocessableEntityError{Message: "you can't review some content more than one time"}
		}

		if input.EpisodeID != nil && review.EpisodeID == *input.EpisodeID {
			return nil, &apperror.UnprocessableEntityError{Message: "you can't review some content more than one time"}
		}
	}

	reviewID := uuid.New()
	params := sqlc.CreateReviewParams{
		ID:     reviewID,
		Rating: input.Rating,
	}

	params.ProfileID = profileID

	if input.MovieID != nil {
		params.MovieID = pgtype.UUID{Bytes: *input.MovieID, Valid: true}
	}

	if input.EpisodeID != nil {
		params.EpisodeID = pgtype.UUID{Bytes: *input.EpisodeID, Valid: true}
	}

	if input.Comment != nil && strings.TrimSpace(*input.Comment) != "" {
		params.Comment = input.Comment
	}

	r, err := s.repo.CreateReview(ctx, params)
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			return nil, &apperror.ConflictError{Field: "review + (episode or movie)"}
		}
		return nil, fmt.Errorf("failed to insert review on database: %w", err)
	}

	return toGraphQLModel(r), nil
}

func (s *ServiceImpl) GetReview(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	r, err := s.repo.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "review"}
		}
		return nil, fmt.Errorf("failed to fetch review %v from database: %w", id, err)
	}

	return toGraphQLModel(r), nil
}

func (s *ServiceImpl) ListReviewsByProfile(ctx context.Context, profileID uuid.UUID) ([]*model.Review, error) {
	reviews, err := s.repo.ListReviewsByProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all reviews from database: %w", err)
	}

	result := make([]*model.Review, len(reviews))
	for i, r := range reviews {
		result[i] = toGraphQLModel(r)
	}
	return result, nil
}

func (s *ServiceImpl) ListReviewsByEpisode(ctx context.Context, episodeID uuid.UUID, profileID uuid.UUID, userID uuid.UUID) ([]*model.Review, error) {
	_, err := s.episodeService.GetEpisode(ctx, episodeID, profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episode %v from database: %w", episodeID, err)
	}

	episodeIDPG := pgtype.UUID{Bytes: episodeID, Valid: true}
	reviews, err := s.repo.ListReviewsByEpisode(ctx, episodeIDPG)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all reviews from database: %w", err)
	}

	result := make([]*model.Review, len(reviews))
	for i, r := range reviews {
		result[i] = toGraphQLModel(r)
	}

	return result, nil
}

func (s *ServiceImpl) ListReviewsByMovie(ctx context.Context, movieID uuid.UUID) ([]*model.Review, error) {
	_, err := s.repo.GetMovie(ctx, movieID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "movie"}
		}
		return nil, fmt.Errorf("failed to fetch movie %v from database: %w", movieID, err)
	}

	movieIDPG := pgtype.UUID{Bytes: movieID, Valid: true}
	reviews, err := s.repo.ListReviewsByMovie(ctx, movieIDPG)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all movie %v reviews from database: %w", movieID, err)
	}

	result := make([]*model.Review, len(reviews))
	for i, r := range reviews {
		result[i] = toGraphQLModel(r)
	}

	return result, nil
}

func (s *ServiceImpl) UpdateReview(ctx context.Context, id uuid.UUID, input model.UpdateReviewInput, profileID uuid.UUID) (*model.Review, error) {
	if input.Rating != nil {
		if _, err := NewRating(*input.Rating); err != nil {
			return nil, &apperror.ValidationError{Field: "rating", Message: err.Error()}
		}
	}

	current, err := s.repo.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "review"}
		}
		return nil, fmt.Errorf("failed to get review %v to update from database: %w", id, err)
	}

	entity := toEntity(current)
	if !entity.BelongsTo(profileID) {
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
		params.Comment = input.Comment
	}

	r, err := s.repo.UpdateReview(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update review %v from database: %w", id, err)
	}

	return toGraphQLModel(r), nil
}

func (s *ServiceImpl) DeleteReview(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error {
	current, err := s.repo.GetReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "review"}
		}
		return fmt.Errorf("failed to get review %v to update from database: %w", id, err)
	}

	entity := toEntity(current)
	if !entity.BelongsTo(profileID) {
		return &apperror.ForbiddenError{Message: "you can't delete reviews that's not yours"}
	}

	if err := s.repo.DeleteReview(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "review"}
		}
		return fmt.Errorf("failed to delete review %v from database: %w", id, err)
	}
	return nil
}
