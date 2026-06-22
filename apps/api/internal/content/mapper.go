package content

import (
	"fmt"
	"strings"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
)

func parseContentStatus(s sqlc.ContentStatus) (*model.ContentStatus, error) {
	status := model.ContentStatus(s)

	switch status {
	case model.ContentStatusPending, model.ContentStatusProcessed:
		return &status, nil
	default:
		return nil, fmt.Errorf("invalid content status: %s", s)
	}
}

func toGraphQlModel(c sqlc.Content, contentURL *string, durationSeconds *int32, status *model.ContentStatus) *model.Content {
	return &model.Content{
		ID:              c.ID,
		Title:           c.Title,
		Description:     c.Description,
		MaturityRating:  model.MaturityRating(c.MaturityRating),
		ContentType:     model.ContentType(c.ContentType),
		ReleaseDate:     c.ReleaseDate.String(),
		GenreID:         c.GenreID,
		ContentURL:      contentURL,
		DurationSeconds: durationSeconds,
		Status:          status,
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
