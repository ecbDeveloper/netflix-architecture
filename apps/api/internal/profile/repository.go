package profile

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	CreateProfile(ctx context.Context, params sqlc.CreateProfileParams) (sqlc.Profile, error)
	GetProfile(ctx context.Context, id uuid.UUID) (sqlc.Profile, error)
	ListProfilesByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Profile, error)
	UpdateProfile(ctx context.Context, params sqlc.UpdateProfileParams) (sqlc.Profile, error)
	DeleteProfile(ctx context.Context, id uuid.UUID) error
}
