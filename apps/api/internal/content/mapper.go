package content

import (
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func toGraphQlModel(c sqlc.Content, contentURL *string, durationMinutes *int32) *model.Content {
	return &model.Content{
		ID:              c.ID,
		Title:           c.Title,
		Description:     c.Description,
		MaturityRating:  model.MaturityRating(c.MaturityRating),
		ContentType:     model.ContentType(c.ContentType),
		ReleaseDate:     c.ReleaseDate.String(),
		GenreID:         c.GenreID,
		ContentURL:      contentURL,
		DurationMinutes: durationMinutes,
		CreatedAt:       c.CreatedAt.String(),
		UpdatedAt:       c.UpdatedAt.String(),
	}
}

func toContentEntity(c sqlc.Content) Content {
	return Content{
		ID:             c.ID,
		Title:          c.Title,
		Description:    c.Description,
		MaturityRating: MaturityRating(c.MaturityRating),
		ContentType:    ContentType(c.ContentType),
		GenreID:        c.GenreID,
	}
}

func graphQLToDBMaturityRating(maturityRating model.MaturityRating) sqlc.MaturityRating {
	prefix := shared.MaturityRatingPrefix
	normalizedMaturityRating := strings.TrimPrefix(maturityRating.String(), prefix)

	return sqlc.MaturityRating(normalizedMaturityRating)
}
