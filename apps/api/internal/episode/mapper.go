package episode

import (
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
)

func toGraphQLModel(e sqlc.Episode) *model.Episode {
	return &model.Episode{
		ID:              e.ID,
		SeriesID:        e.SeriesID,
		Season:          e.Season,
		EpisodeNumber:   e.EpisodeNumber,
		Title:           e.Title,
		DurationMinutes: e.DurationMinutes,
		CreatedAt:       e.CreatedAt.String(),
	}
}

func toEpisodeEntity(e sqlc.Episode) Episode {
	return Episode{
		ID:       e.ID,
		SeriesID: e.SeriesID,
		Title:    e.Title,
	}
}
