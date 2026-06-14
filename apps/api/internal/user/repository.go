package user

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	ListUsers(ctx context.Context) ([]sqlc.User, error)
	UpdateUser(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
