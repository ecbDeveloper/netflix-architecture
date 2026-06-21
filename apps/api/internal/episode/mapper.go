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

func pgInt4ToInt32Ptr(t pgtype.Int4) *int32 {
	if !t.Valid {
		return nil
	}
	i := t.Int32
	return &i
}

func toGraphQLModel(e sqlc.Episode, contentURL *string, durationSeconds *int32) *model.Episode {
	return &model.Episode{
		ID:              e.ID,
		SeriesID:        e.SeriesID,
		Season:          e.Season,
		EpisodeNumber:   e.EpisodeNumber,
		Title:           e.Title,
		DurationSeconds: durationSeconds,
		ContentURL:      contentURL,
		Status:          model.ContentStatus(e.Status),
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
