package infra

import (
	"context"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}
