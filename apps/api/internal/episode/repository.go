package episode

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	CreateEpisode(ctx context.Context, params sqlc.CreateEpisodeParams) (sqlc.Episode, error)
	GetEpisode(ctx context.Context, id uuid.UUID) (sqlc.Episode, error)
	GetSeries(ctx context.Context, id uuid.UUID) (sqlc.GetSeriesRow, error)
	ListEpisodesBySeries(ctx context.Context, seriesID uuid.UUID) ([]sqlc.Episode, error)
	UpdateEpisode(ctx context.Context, params sqlc.UpdateEpisodeParams) (sqlc.Episode, error)
	DeleteEpisode(ctx context.Context, id uuid.UUID) error
}
