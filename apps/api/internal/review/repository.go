package review

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	CreateReview(ctx context.Context, params sqlc.CreateReviewParams) (sqlc.Review, error)
	GetReview(ctx context.Context, id uuid.UUID) (sqlc.Review, error)
	GetMovie(ctx context.Context, id uuid.UUID) (sqlc.GetMovieRow, error)
	ListReviewsByProfile(ctx context.Context, profileID uuid.UUID) ([]sqlc.Review, error)
	ListReviewsByEpisode(ctx context.Context, episodeID pgtype.UUID) ([]sqlc.Review, error)
	ListReviewsByMovie(ctx context.Context, movieID pgtype.UUID) ([]sqlc.Review, error)
	UpdateReview(ctx context.Context, params sqlc.UpdateReviewParams) (sqlc.Review, error)
	DeleteReview(ctx context.Context, id uuid.UUID) error
}
