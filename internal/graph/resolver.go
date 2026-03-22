package graph

import (
	"log/slog"

	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Queries *sqlc.Queries
	Logger  *slog.Logger
}
