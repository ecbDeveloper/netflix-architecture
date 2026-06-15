package review

import (
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/google/uuid"
)

func toGraphQLModel(r sqlc.Review) *model.Review {
	m := &model.Review{
		ID:        r.ID,
		Rating:    r.Rating,
		CreatedAt: r.CreatedAt.String(),
		UpdatedAt: r.UpdatedAt.String(),
	}

	if r.EpisodeID.Valid {
		m.EpisodeID = r.EpisodeID.Bytes
	}
	if r.MovieID.Valid {
		m.MovieID = r.MovieID.Bytes
	}

	if r.Comment.Valid {
		m.Comment = &r.Comment.String
	}

	return m
}

func toEntity(r sqlc.Review) Review {
	entity := Review{
		ID:        r.ID,
		ProfileID: r.ProfileID,
	}
	if r.MovieID.Valid {
		id := uuid.UUID(r.MovieID.Bytes)
		entity.MovieID = &id
	}
	if r.EpisodeID.Valid {
		id := uuid.UUID(r.EpisodeID.Bytes)
		entity.EpisodeID = &id
	}
	return entity
}
