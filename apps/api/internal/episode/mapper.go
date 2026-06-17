package episode

import (
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func toGraphQLModel(e sqlc.Episode, contentURL *string) *model.Episode {
	return &model.Episode{
		ID:              e.ID,
		SeriesID:        e.SeriesID,
		Season:          e.Season,
		EpisodeNumber:   e.EpisodeNumber,
		Title:           e.Title,
		DurationMinutes: e.DurationMinutes,
		ContentURL:      contentURL,
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
