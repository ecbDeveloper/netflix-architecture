package profile

import (
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
)

func toGraphQLModel(p sqlc.Profile) *model.Profile {
	return &model.Profile{
		ID:                  p.ID,
		UserID:              p.UserID,
		Name:                p.Name,
		HasParentalControls: p.HasParentalControls,
		CreatedAt:           p.CreatedAt.String(),
		UpdatedAt:           p.UpdatedAt.String(),
	}
}

func toEntity(p sqlc.Profile) Profile {
	return Profile{
		ID:                  p.ID,
		UserID:              p.UserID,
		Name:                p.Name,
		HasParentalControls: p.HasParentalControls,
	}
}
