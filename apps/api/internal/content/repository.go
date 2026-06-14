package content

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	CreateContent(ctx context.Context, params sqlc.CreateContentParams) error
	GetContent(ctx context.Context, id uuid.UUID) (sqlc.Content, error)
	ListContents(ctx context.Context) ([]sqlc.Content, error)
	ListKidsContents(ctx context.Context) ([]sqlc.Content, error)
	ListContentsByType(ctx context.Context, contentType sqlc.ContentType) ([]sqlc.Content, error)
	ListKidsContentsByType(ctx context.Context, contentType sqlc.ContentType) ([]sqlc.Content, error)
	ListContentsByGenre(ctx context.Context, genreID int32) ([]sqlc.Content, error)
	ListKidsContentsByGenre(ctx context.Context, genreID int32) ([]sqlc.Content, error)
	UpdateContent(ctx context.Context, params sqlc.UpdateContentParams) (sqlc.Content, error)
	DeleteContent(ctx context.Context, id uuid.UUID) error

	CreateMovie(ctx context.Context, params sqlc.CreateMovieParams) (sqlc.Movie, error)
	GetMovie(ctx context.Context, contentID uuid.UUID) (sqlc.GetMovieRow, error)
	UpdateMovie(ctx context.Context, params sqlc.UpdateMovieParams) (sqlc.Movie, error)

	CreateSeries(ctx context.Context, contentID uuid.UUID) (uuid.UUID, error)

	ListEpisodesBySeries(ctx context.Context, seriesID uuid.UUID) ([]sqlc.Episode, error)

	ListContentGenres(ctx context.Context) ([]sqlc.ContentGenre, error)
}
